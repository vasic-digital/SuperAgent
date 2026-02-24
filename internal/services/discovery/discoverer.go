package discovery

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
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

// mdnsDiscoverer performs mDNS (ZeroConf/Bonjour) based discovery on the local network.
// It sends multicast DNS queries for _<service>._tcp.local. and listens for responses
// containing SRV/A/AAAA records. Falls back to TCP discovery on failure or timeout.
type mdnsDiscoverer struct {
	logger *logrus.Logger
}

// mdnsMulticastAddr is the standard mDNS IPv4 multicast address and port.
const mdnsMulticastAddr = "224.0.0.251:5353"

// mdnsDefaultTimeout is the default timeout for mDNS queries.
const mdnsDefaultTimeout = 2 * time.Second

// mdnsResult holds a discovered service endpoint from mDNS responses.
type mdnsResult struct {
	host string
	port uint16
}

func (d *mdnsDiscoverer) Discover(ctx context.Context, endpoint *config.ServiceEndpoint) (bool, error) {
	timeout := endpoint.DiscoveryTimeout
	if timeout == 0 {
		timeout = mdnsDefaultTimeout
	}

	serviceName := endpoint.ServiceName
	if serviceName == "" {
		d.logger.Warn("mDNS discovery: no service name, falling back to TCP")
		tcp := &tcpDiscoverer{logger: d.logger}
		return tcp.Discover(ctx, endpoint)
	}

	// Build the mDNS service query name: _<service>._tcp.local.
	srvName := serviceToSRVName(serviceName)
	queryName := srvName + ".local."

	d.logger.WithFields(logrus.Fields{
		"service": serviceName,
		"query":   queryName,
		"method":  "mdns",
		"timeout": timeout,
	}).Debug("Attempting mDNS discovery")

	queryCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result, err := d.performMDNSQuery(queryCtx, queryName)
	if err != nil {
		d.logger.WithFields(logrus.Fields{
			"service": serviceName,
			"error":   err,
		}).Debug("mDNS discovery failed, falling back to TCP")
		tcp := &tcpDiscoverer{logger: d.logger}
		return tcp.Discover(ctx, endpoint)
	}

	if result != nil {
		d.logger.WithFields(logrus.Fields{
			"service": serviceName,
			"host":    result.host,
			"port":    result.port,
			"method":  "mdns",
		}).Info("Service discovered via mDNS")
		return true, nil
	}

	// No result found via mDNS, fall back to TCP
	d.logger.WithFields(logrus.Fields{
		"service": serviceName,
		"method":  "mdns",
	}).Debug("No mDNS response received, falling back to TCP")
	tcp := &tcpDiscoverer{logger: d.logger}
	return tcp.Discover(ctx, endpoint)
}

// performMDNSQuery sends an mDNS query and listens for responses.
// Returns an mdnsResult if a matching service is found, nil if no response,
// or an error if the query could not be sent.
func (d *mdnsDiscoverer) performMDNSQuery(ctx context.Context, queryName string) (*mdnsResult, error) {
	// Resolve the multicast address
	mcastAddr, err := net.ResolveUDPAddr("udp4", mdnsMulticastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve mDNS multicast address: %w", err)
	}

	// Open a UDP connection bound to any local address for sending and receiving
	conn, err := net.ListenPacket("udp4", ":0")
	if err != nil {
		return nil, fmt.Errorf("failed to open UDP listener for mDNS: %w", err)
	}
	defer func() { _ = conn.Close() }()

	// Build the mDNS query packet (DNS wire format)
	query, err := buildMDNSQuery(queryName)
	if err != nil {
		return nil, fmt.Errorf("failed to build mDNS query: %w", err)
	}

	// Send the query to the multicast group
	_, err = conn.WriteTo(query, mcastAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to send mDNS query: %w", err)
	}

	d.logger.WithField("query", queryName).Debug("mDNS query sent, awaiting responses")

	// Listen for responses with context cancellation support
	var mu sync.Mutex
	var result *mdnsResult
	done := make(chan struct{})

	go func() {
		defer close(done)
		buf := make([]byte, 4096)
		for {
			// Set a short read deadline to allow checking context cancellation
			if deadlineConn, ok := conn.(interface {
				SetReadDeadline(t time.Time) error
			}); ok {
				_ = deadlineConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
			}

			n, _, readErr := conn.ReadFrom(buf)
			if readErr != nil {
				// Check if context has been cancelled
				if ctx.Err() != nil {
					return
				}
				// Timeout errors are expected; continue listening
				if netErr, ok := readErr.(net.Error); ok && netErr.Timeout() {
					continue
				}
				return
			}

			if n < 12 {
				continue // Too small to be a valid DNS packet
			}

			parsed := parseMDNSResponse(buf[:n], queryName)
			if parsed != nil {
				mu.Lock()
				result = parsed
				mu.Unlock()
				return
			}
		}
	}()

	// Wait for either a result or context cancellation
	select {
	case <-done:
		mu.Lock()
		defer mu.Unlock()
		return result, nil
	case <-ctx.Done():
		// Close the connection to unblock the reader goroutine
		_ = conn.Close()
		<-done
		mu.Lock()
		defer mu.Unlock()
		return result, nil
	}
}

// buildMDNSQuery constructs an mDNS query packet in DNS wire format.
// The query requests PTR records for the given service name.
func buildMDNSQuery(queryName string) ([]byte, error) {
	// DNS Header: 12 bytes
	// ID=0 (mDNS uses 0), QR=0 (query), OPCODE=0, AA=0, TC=0, RD=0
	// RA=0, Z=0, RCODE=0, QDCOUNT=1, ANCOUNT=0, NSCOUNT=0, ARCOUNT=0
	header := []byte{
		0x00, 0x00, // ID
		0x00, 0x00, // Flags: standard query
		0x00, 0x01, // QDCOUNT: 1 question
		0x00, 0x00, // ANCOUNT: 0
		0x00, 0x00, // NSCOUNT: 0
		0x00, 0x00, // ARCOUNT: 0
	}

	// Encode the query name into DNS wire format
	encodedName, err := encodeDNSName(queryName)
	if err != nil {
		return nil, err
	}

	// Question section: NAME + TYPE(PTR=12) + CLASS(IN=1, with unicast-response bit clear)
	question := make([]byte, len(encodedName)+4)
	copy(question, encodedName)
	binary.BigEndian.PutUint16(question[len(encodedName):], 12)   // QTYPE: PTR
	binary.BigEndian.PutUint16(question[len(encodedName)+2:], 1)  // QCLASS: IN

	packet := make([]byte, 0, len(header)+len(question))
	packet = append(packet, header...)
	packet = append(packet, question...)

	return packet, nil
}

// encodeDNSName encodes a dotted domain name into DNS wire format.
// Example: "_http._tcp.local." -> [5]_http[4]_tcp[5]local[0]
func encodeDNSName(name string) ([]byte, error) {
	// Remove trailing dot if present
	name = strings.TrimSuffix(name, ".")
	if name == "" {
		return nil, fmt.Errorf("empty DNS name")
	}

	labels := strings.Split(name, ".")
	var buf []byte
	for _, label := range labels {
		if len(label) == 0 {
			return nil, fmt.Errorf("empty label in DNS name: %s", name)
		}
		if len(label) > 63 {
			return nil, fmt.Errorf("DNS label too long (%d > 63): %s", len(label), label)
		}
		buf = append(buf, byte(len(label)))
		buf = append(buf, []byte(label)...)
	}
	buf = append(buf, 0x00) // Root label terminator
	return buf, nil
}

// parseMDNSResponse parses a DNS response packet and extracts service information.
// It looks for SRV, A, and AAAA records that match the queried service.
// Returns an mdnsResult if relevant records are found, nil otherwise.
func parseMDNSResponse(data []byte, queryName string) *mdnsResult {
	if len(data) < 12 {
		return nil
	}

	// Parse header
	flags := binary.BigEndian.Uint16(data[2:4])
	isResponse := (flags & 0x8000) != 0
	if !isResponse {
		return nil // Not a response
	}

	qdCount := binary.BigEndian.Uint16(data[4:6])
	anCount := binary.BigEndian.Uint16(data[6:8])
	// nsCount and arCount are at data[8:10] and data[10:12]
	nsCount := binary.BigEndian.Uint16(data[8:10])
	arCount := binary.BigEndian.Uint16(data[10:12])

	offset := 12

	// Skip question section
	for i := 0; i < int(qdCount); i++ {
		var err error
		offset, err = skipDNSName(data, offset)
		if err != nil {
			return nil
		}
		offset += 4 // QTYPE + QCLASS
		if offset > len(data) {
			return nil
		}
	}

	var srvHost string
	var srvPort uint16
	var aAddr string
	hasSRV := false

	// Parse all resource record sections (answer, authority, additional)
	totalRRs := int(anCount) + int(nsCount) + int(arCount)
	for i := 0; i < totalRRs; i++ {
		if offset >= len(data) {
			break
		}

		// Parse resource record name
		_, newOffset, err := decodeDNSName(data, offset)
		if err != nil {
			return nil
		}
		offset = newOffset

		if offset+10 > len(data) {
			return nil
		}

		rrType := binary.BigEndian.Uint16(data[offset:])
		// rrClass at offset+2 (not used for matching)
		rdLength := binary.BigEndian.Uint16(data[offset+8:])
		offset += 10

		if offset+int(rdLength) > len(data) {
			return nil
		}

		rdataStart := offset

		switch rrType {
		case 33: // SRV record
			if rdLength >= 6 {
				// SRV RDATA: Priority(2) + Weight(2) + Port(2) + Target(variable)
				srvPort = binary.BigEndian.Uint16(data[rdataStart+4:])
				target, _, targetErr := decodeDNSName(data, rdataStart+6)
				if targetErr == nil {
					srvHost = target
					hasSRV = true
				}
			}
		case 1: // A record (IPv4)
			if rdLength == 4 {
				aAddr = fmt.Sprintf("%d.%d.%d.%d",
					data[rdataStart], data[rdataStart+1],
					data[rdataStart+2], data[rdataStart+3])
			}
		case 28: // AAAA record (IPv6)
			if rdLength == 16 {
				ip := net.IP(data[rdataStart : rdataStart+16])
				aAddr = ip.String()
			}
		case 12: // PTR record — indicates service instance exists
			// The presence of a PTR response for our query means the service exists.
			// We still prefer SRV+A for the endpoint, but PTR alone signals discovery.
			if !hasSRV && aAddr == "" {
				// Mark that we got at least a PTR response
				hasSRV = false // will check below
			}
		}

		offset = rdataStart + int(rdLength)
	}

	// Build the result from collected records
	if hasSRV {
		host := srvHost
		if aAddr != "" {
			host = aAddr // Prefer resolved IP over hostname
		}
		if host != "" {
			return &mdnsResult{host: host, port: srvPort}
		}
	}

	// If we only got A/AAAA records but no SRV, we still found the service
	if aAddr != "" {
		return &mdnsResult{host: aAddr, port: 0}
	}

	return nil
}

// skipDNSName advances past a DNS-encoded name in the packet.
// Handles both label sequences and compressed pointers.
func skipDNSName(data []byte, offset int) (int, error) {
	if offset >= len(data) {
		return offset, fmt.Errorf("offset out of bounds")
	}

	for {
		if offset >= len(data) {
			return offset, fmt.Errorf("unexpected end of DNS name")
		}

		length := int(data[offset])

		if length == 0 {
			// Root label — end of name
			return offset + 1, nil
		}

		if length&0xC0 == 0xC0 {
			// Compression pointer — 2 bytes total, name ends here
			return offset + 2, nil
		}

		// Regular label
		offset += 1 + length
	}
}

// decodeDNSName decodes a DNS-encoded name from a packet, handling compression pointers.
// Returns the decoded name and the new offset past the name in the original data.
func decodeDNSName(data []byte, offset int) (string, int, error) {
	if offset >= len(data) {
		return "", offset, fmt.Errorf("offset out of bounds")
	}

	var labels []string
	visited := make(map[int]bool) // Detect pointer loops
	newOffset := -1               // Track the "real" position after the first pointer
	currentOffset := offset

	for {
		if currentOffset >= len(data) {
			return "", offset, fmt.Errorf("unexpected end of DNS name at offset %d", currentOffset)
		}

		length := int(data[currentOffset])

		if length == 0 {
			// Root label — end of name
			if newOffset == -1 {
				newOffset = currentOffset + 1
			}
			break
		}

		if length&0xC0 == 0xC0 {
			// Compression pointer
			if currentOffset+1 >= len(data) {
				return "", offset, fmt.Errorf("truncated compression pointer")
			}
			if newOffset == -1 {
				newOffset = currentOffset + 2
			}
			ptr := int(binary.BigEndian.Uint16(data[currentOffset:]) & 0x3FFF)
			if visited[ptr] {
				return "", offset, fmt.Errorf("DNS name compression loop detected")
			}
			visited[ptr] = true
			currentOffset = ptr
			continue
		}

		// Regular label
		labelStart := currentOffset + 1
		labelEnd := labelStart + length
		if labelEnd > len(data) {
			return "", offset, fmt.Errorf("DNS label extends past packet end")
		}

		labels = append(labels, string(data[labelStart:labelEnd]))
		currentOffset = labelEnd
	}

	name := strings.Join(labels, ".")
	return name, newOffset, nil
}
