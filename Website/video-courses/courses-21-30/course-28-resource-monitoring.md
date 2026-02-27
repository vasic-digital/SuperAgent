# Course-28: Resource Monitoring

## Course Information
- **Duration:** 45 minutes
- **Level:** Advanced
- **Prerequisites:** Course-27

## Module 1: Resource Types (8 min)

**What to Monitor:**
- CPU usage
- Memory consumption
- Goroutine count
- File descriptors
- Network I/O
- Disk usage

## Module 2: Go Runtime Metrics (12 min)

**Runtime Package:**
```go
var m runtime.MemStats
runtime.ReadMemStats(&m)

fmt.Printf("Alloc: %d MB\n", m.Alloc/1024/1024)
fmt.Printf("Goroutines: %d\n", runtime.NumGoroutine())
fmt.Printf("CPU Count: %d\n", runtime.NumCPU())
```

**Prometheus Integration:**
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    memAlloc = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "go_memory_alloc_bytes",
        Help: "Current memory allocation",
    })
)

func init() {
    prometheus.MustRegister(memAlloc)
}

func updateMetrics() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    memAlloc.Set(float64(m.Alloc))
}
```

## Module 3: Custom Metrics (15 min)

**Business Metrics:**
```go
var (
    requestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Request duration",
        },
        []string{"method", "endpoint"},
    )
)

func handler(w http.ResponseWriter, r *http.Request) {
    start := time.Now()
    defer func() {
        requestDuration.WithLabelValues(
            r.Method, r.URL.Path,
        ).Observe(time.Since(start).Seconds())
    }()
    
    // Handler logic
}
```

**Health Checks:**
```go
func healthCheck(w http.ResponseWriter, r *http.Request) {
    status := map[string]interface{}{
        "status": "healthy",
        "checks": map[string]bool{
            "database": checkDB(),
            "cache": checkCache(),
            "external": checkExternal(),
        },
    }
    
    json.NewEncoder(w).Encode(status)
}
```

## Module 4: Alerting (10 min)

**Prometheus Alerting Rules:**
```yaml
groups:
  - name: helixagent
    rules:
      - alert: HighMemoryUsage
        expr: go_memory_alloc_bytes > 1e9
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: High memory usage detected
```

**Grafana Dashboards:**
- Import dashboard ID: 12345
- Configure data source: Prometheus
- Set up alerts

## Assessment

**Lab:** Implement monitoring for a service.
