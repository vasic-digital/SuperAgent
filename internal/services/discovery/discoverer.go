package discovery

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"dev.helix.agent/internal/config"
	"github.com/sirupsen/logrus"
)

// Discoverer defines the interface for service discovery implementations.
type Discoverer interface {
	// Discover attempts to discover a service endpoint in the network.
	// Returns true if the service is discovered and reachable.
	Discover(ctx context.Context, endpoint *config.ServiceEndpoint) (bool, error)
}

// NewDiscoverer creates a Discoverer based on the endpoint's discovery method.
// Defaults to TCP discovery if method is empty or unsupported.
func NewDiscoverer(logger *logrus.Logger) Discoverer {
	return &compositeDiscoverer{
		tcp:  &tcpDiscoverer{logger: logger},
		http: &httpDiscoverer{logger: logger},
		dns:  &dnsDiscoverer{logger: logger, resolver: &defaultDNSResolver{}},
		mdns: &mdnsDiscoverer{logger: logger},
	}
}

// compositeDiscoverer delegates to appropriate discoverer based on method.
type compositeDiscoverer struct {
	tcp  *tcpDiscoverer
	http *httpDiscoverer
	dns  *dnsDiscoverer
	mdns *mdnsDiscoverer
}

func (c *compositeDiscoverer) Discover(ctx context.Context, endpoint *config.ServiceEndpoint) (bool, error) {
	method := endpoint.DiscoveryMethod
	if method == "" {
		method = "tcp"
	}

	switch strings.ToLower(method) {
	case "tcp":
		return c.tcp.Discover(ctx, endpoint)
	case "http":
		return c.http.Discover(ctx, endpoint)
	case "dns":
		return c.dns.Discover(ctx, endpoint)
	case "mdns":
		return c.mdns.Discover(ctx, endpoint)
	default:
		// Fallback to TCP
		return c.tcp.Discover(ctx, endpoint)
	}
}

// tcpDiscoverer performs TCP connection-based discovery.
type tcpDiscoverer struct {
	logger *logrus.Logger
}

func (d *tcpDiscoverer) Discover(ctx context.Context, endpoint *config.ServiceEndpoint) (bool, error) {
	timeout := endpoint.DiscoveryTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	addr := endpoint.ResolvedURL()
	if addr == "" {
		return false, fmt.Errorf("no address configured for service")
	}

	d.logger.WithFields(logrus.Fields{
		"service": endpoint.ServiceName,
		"address": addr,
		"method":  "tcp",
	}).Debug("Attempting TCP discovery")

	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return false, nil // Not discovered, but not an error
	}
	_ = conn.Close()
	return true, nil
}

// httpDiscoverer performs HTTP-based discovery using health endpoint.
type httpDiscoverer struct {
	logger *logrus.Logger
}

func (d *httpDiscoverer) Discover(ctx context.Context, endpoint *config.ServiceEndpoint) (bool, error) {
	timeout := endpoint.DiscoveryTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	baseURL := endpoint.ResolvedURL()
	if baseURL == "" {
		return false, fmt.Errorf("no address configured for service")
	}

	// Build URL similar to health check
	url := baseURL
	if endpoint.HealthPath != "" {
		if len(url) > 0 && url[0] != 'h' {
			url = "http://" + url
		}
		url = url + endpoint.HealthPath
	} else {
		if len(url) > 0 && url[0] != 'h' {
			url = "http://" + url
		}
	}

	d.logger.WithFields(logrus.Fields{
		"service": endpoint.ServiceName,
		"url":     url,
		"method":  "http",
	}).Debug("Attempting HTTP discovery")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %w", err)
	}

	client := &http.Client{Timeout: timeout}
	resp, err := client.Do(req)
	if err != nil {
		return false, nil // Not discovered
	}
	_ = resp.Body.Close()

	// Consider any 2xx or 3xx status as discovered
	if resp.StatusCode >= 200 && resp.StatusCode < 400 {
		return true, nil
	}
	return false, nil
}

// dnsDiscoverer performs DNS-based discovery (SRV or A records).
// dnsResolver defines the interface for DNS lookups (allows mocking in tests)
type dnsResolver interface {
	LookupSRV(ctx context.Context, service, proto, name string) (string, []*net.SRV, error)
	LookupHost(ctx context.Context, host string) ([]string, error)
}

// defaultDNSResolver wraps net.DefaultResolver to implement dnsResolver interface.
type defaultDNSResolver struct{}

func (r *defaultDNSResolver) LookupSRV(ctx context.Context, service, proto, name string) (string, []*net.SRV, error) {
	return net.DefaultResolver.LookupSRV(ctx, service, proto, name)
}

func (r *defaultDNSResolver) LookupHost(ctx context.Context, host string) ([]string, error) {
	return net.DefaultResolver.LookupHost(ctx, host)
}

// dnsDiscoverer performs DNS-based discovery (SRV or A records).
type dnsDiscoverer struct {
	logger   *logrus.Logger
	resolver dnsResolver
}

// isIPAddress checks if a string is an IP address (IPv4 or IPv6).
func isIPAddress(host string) bool {
	return net.ParseIP(host) != nil
}

// serviceToSRVName maps service names to SRV service names.
// Returns empty string if no mapping.
func serviceToSRVName(serviceName string) string {
	// Map common service names to SRV service names
	mapping := map[string]string{
		"postgresql": "_postgresql._tcp",
		"redis":      "_redis._tcp",
		"chromadb":   "_chroma._tcp",
		"cognee":     "_cognee._tcp",
		"prometheus": "_prometheus._tcp",
		"grafana":    "_grafana._tcp",
		"neo4j":      "_neo4j._tcp",
		"kafka":      "_kafka._tcp",
		"rabbitmq":   "_rabbitmq._tcp",
		"qdrant":     "_qdrant._tcp",
		"weaviate":   "_weaviate._tcp",
		"langchain":  "_langchain._tcp",
		"llamaindex": "_llamaindex._tcp",
		"mcpservers": "_mcp._tcp",
	}
	if srv, ok := mapping[strings.ToLower(serviceName)]; ok {
		return srv
	}
	// Default pattern: _service._tcp
	return "_" + strings.ToLower(serviceName) + "._tcp"
}

func (d *dnsDiscoverer) Discover(ctx context.Context, endpoint *config.ServiceEndpoint) (bool, error) {
	timeout := endpoint.DiscoveryTimeout
	if timeout == 0 {
		timeout = 5 * time.Second
	}

	host := endpoint.Host
	if host == "" {
		return false, fmt.Errorf("no host configured for service")
	}

	d.logger.WithFields(logrus.Fields{
		"service": endpoint.ServiceName,
		"host":    host,
		"method":  "dns",
	}).Debug("Attempting DNS discovery")

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// If host is already an IP address, skip DNS resolution
	if isIPAddress(host) {
		d.logger.WithField("host", host).Debug("Host is an IP address, DNS resolution skipped")
		// Still need to check if service is reachable via TCP
		tcp := &tcpDiscoverer{logger: d.logger}
		return tcp.Discover(ctx, endpoint)
	}

	// Try SRV discovery first
	srvName := serviceToSRVName(endpoint.ServiceName)
	fullSRVName := srvName + "." + host

	d.logger.WithField("srv", fullSRVName).Debug("Attempting SRV lookup")
	_, srvAddrs, err := d.resolver.LookupSRV(ctx, "", "", fullSRVName)
	if err == nil && len(srvAddrs) > 0 {
		d.logger.WithFields(logrus.Fields{
			"service": endpoint.ServiceName,
			"srv":     fullSRVName,
			"targets": len(srvAddrs),
		}).Debug("SRV records found")
		// SRV records exist, service is discoverable via DNS
		// Optionally we could try connecting to one of the targets,
		// but for discovery purposes, SRV records indicate service availability
		return true, nil
	}

	// Fall back to A/AAAA record lookup
	d.logger.WithField("host", host).Debug("Attempting A/AAAA lookup")
	hostAddrs, err := d.resolver.LookupHost(ctx, host)
	if err != nil {
		d.logger.WithField("host", host).Debug("DNS lookup failed")
		return false, nil // Not discovered
	}

	d.logger.WithFields(logrus.Fields{
		"service": endpoint.ServiceName,
		"host":    host,
		"ips":     len(hostAddrs),
	}).Debug("DNS resolution successful")
	// DNS resolution succeeded, service hostname is resolvable
	return true, nil
}

// mdnsDiscoverer performs mDNS (ZeroConf/Bonjour) based discovery.
type mdnsDiscoverer struct {
	logger *logrus.Logger
}

func (d *mdnsDiscoverer) Discover(ctx context.Context, endpoint *config.ServiceEndpoint) (bool, error) {
	d.logger.WithField("service", endpoint.ServiceName).Warn("mDNS discovery not yet implemented, falling back to TCP")
	// Delegate to TCP discoverer for now
	tcp := &tcpDiscoverer{logger: d.logger}
	return tcp.Discover(ctx, endpoint)
}
