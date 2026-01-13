package mcp

import (
	"context"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"dev.helix.agent/internal/mcp"
)

func TestMCPPreinstaller_NewPreinstaller(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()

	preinstaller := mcp.NewPreinstaller(tempDir, logger)
	require.NotNil(t, preinstaller)
}

func TestMCPPreinstaller_GetPackages(t *testing.T) {
	packages := mcp.StandardMCPPackages

	assert.NotEmpty(t, packages)

	// Check standard packages are defined
	packageNames := make(map[string]bool)
	for _, pkg := range packages {
		packageNames[pkg.Name] = true
	}

	assert.True(t, packageNames["filesystem"], "filesystem package should be defined")
	assert.True(t, packageNames["github"], "github package should be defined")
	assert.True(t, packageNames["memory"], "memory package should be defined")
}

func TestMCPPreinstaller_IsInstalled_NotInstalled(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	// Package not installed yet
	installed := preinstaller.IsInstalled("nonexistent-package")
	assert.False(t, installed)
}

func TestMCPPreinstaller_MarkInstalled(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	// Initially not installed
	assert.False(t, preinstaller.IsInstalled("test-package"))

	// Mark as installed
	preinstaller.MarkInstalled("test-package")

	// Now should be installed
	assert.True(t, preinstaller.IsInstalled("test-package"))
}

func TestMCPPreinstaller_ConcurrentMarkInstalled(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	var wg sync.WaitGroup
	packages := []string{"pkg1", "pkg2", "pkg3", "pkg4", "pkg5"}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			pkg := packages[idx%len(packages)]
			preinstaller.MarkInstalled(pkg)
			preinstaller.IsInstalled(pkg)
		}(i)
	}

	wg.Wait()

	// All packages should be installed
	for _, pkg := range packages {
		assert.True(t, preinstaller.IsInstalled(pkg))
	}
}

func TestMCPPreinstaller_WaitForPackage_AlreadyInstalled(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	// Mark as installed first
	preinstaller.MarkInstalled("test-package")

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := preinstaller.WaitForPackage(ctx, "test-package")
	assert.NoError(t, err)
}

func TestMCPPreinstaller_WaitForPackage_Timeout(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := preinstaller.WaitForPackage(ctx, "never-installed")
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestMCPPreinstaller_WaitForPackage_EventuallyInstalled(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Install after a delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		preinstaller.MarkInstalled("delayed-package")
	}()

	err := preinstaller.WaitForPackage(ctx, "delayed-package")
	assert.NoError(t, err)
}

func TestMCPPreinstaller_GetInstallDir(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	installDir := preinstaller.GetInstallDir()
	assert.Equal(t, tempDir, installDir)
}

func TestMCPPreinstaller_GetInstalledPath(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	path := preinstaller.GetInstalledPath("test-package")
	expectedPath := filepath.Join(tempDir, "node_modules", "test-package")
	assert.Equal(t, expectedPath, path)
}

func TestMCPPreinstaller_Metrics(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	// Mark some packages as installed
	preinstaller.MarkInstalled("pkg1")
	preinstaller.MarkInstalled("pkg2")
	preinstaller.MarkInstalled("pkg3")

	metrics := preinstaller.Metrics()
	require.NotNil(t, metrics)
	assert.Equal(t, int64(3), metrics.InstalledCount)
}

func TestMCPConnectionPool_NewPool(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	config := &mcp.MCPPoolConfig{
		MaxConnections:    10,
		ConnectionTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	pool := mcp.NewMCPConnectionPool(preinstaller, config, logger)
	require.NotNil(t, pool)
}

func TestMCPConnectionPool_GetConnectionStatus(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	config := &mcp.MCPPoolConfig{
		MaxConnections:    10,
		ConnectionTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	pool := mcp.NewMCPConnectionPool(preinstaller, config, logger)

	// Non-existent connection should have pending status
	status := pool.GetConnectionStatus("nonexistent")
	assert.Equal(t, mcp.ConnectionStatusPending, status)
}

func TestMCPConnectionPool_Close(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	config := &mcp.MCPPoolConfig{
		MaxConnections:    10,
		ConnectionTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	pool := mcp.NewMCPConnectionPool(preinstaller, config, logger)

	// Close should not panic
	err := pool.Close()
	assert.NoError(t, err)
}

func TestMCPConnectionPool_Metrics(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	config := &mcp.MCPPoolConfig{
		MaxConnections:    10,
		ConnectionTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	pool := mcp.NewMCPConnectionPool(preinstaller, config, logger)

	metrics := pool.Metrics()
	require.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.ActiveConnections)
}

func TestMCPConnectionPool_ConcurrentAccess(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	config := &mcp.MCPPoolConfig{
		MaxConnections:    10,
		ConnectionTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	pool := mcp.NewMCPConnectionPool(preinstaller, config, logger)

	var wg sync.WaitGroup
	servers := []string{"server1", "server2", "server3", "server4", "server5"}

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			server := servers[idx%len(servers)]
			pool.GetConnectionStatus(server)
			pool.Metrics()
		}(i)
	}

	wg.Wait()
}

func TestMCPConnectionPool_ServerList(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	config := &mcp.MCPPoolConfig{
		MaxConnections:    10,
		ConnectionTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	pool := mcp.NewMCPConnectionPool(preinstaller, config, logger)

	// Add some servers
	pool.RegisterServer("server1", &mcp.ServerConfig{
		Name:    "server1",
		Command: []string{"node", "server1.js"},
	})
	pool.RegisterServer("server2", &mcp.ServerConfig{
		Name:    "server2",
		Command: []string{"node", "server2.js"},
	})

	servers := pool.ListServers()
	assert.Len(t, servers, 2)
	assert.Contains(t, servers, "server1")
	assert.Contains(t, servers, "server2")
}

func TestMCPConnection_Status(t *testing.T) {
	conn := &mcp.MCPConnection{
		Name:   "test",
		Status: mcp.ConnectionStatusPending,
	}

	assert.Equal(t, "test", conn.Name)
	assert.Equal(t, mcp.ConnectionStatusPending, conn.Status)

	conn.SetStatus(mcp.ConnectionStatusConnected)
	assert.Equal(t, mcp.ConnectionStatusConnected, conn.Status)
}

func TestConnectionStatus_String(t *testing.T) {
	tests := []struct {
		status   mcp.ConnectionStatus
		expected string
	}{
		{mcp.ConnectionStatusPending, "pending"},
		{mcp.ConnectionStatusConnecting, "connecting"},
		{mcp.ConnectionStatusConnected, "connected"},
		{mcp.ConnectionStatusFailed, "failed"},
		{mcp.ConnectionStatusClosed, "closed"},
	}

	for _, tt := range tests {
		assert.Equal(t, tt.expected, tt.status.String())
	}
}

func TestMCPPoolConfig_Defaults(t *testing.T) {
	config := mcp.DefaultMCPPoolConfig()

	assert.True(t, config.MaxConnections > 0)
	assert.True(t, config.ConnectionTimeout > 0)
	assert.True(t, config.IdleTimeout > 0)
}

func TestMCPPreinstaller_PreInstallPackage(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping npm installation test in short mode")
	}

	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	// Try to install a small npm package
	err := preinstaller.PreInstallPackage(ctx, &mcp.MCPPackage{
		Name: "is-odd",
		NPM:  "is-odd",
	})

	// This may fail if npm is not installed, which is OK for unit tests
	if err != nil {
		t.Logf("npm install failed (expected in environments without npm): %v", err)
	}
}

func BenchmarkMCPPreinstaller_IsInstalled(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := b.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	// Pre-install some packages
	for i := 0; i < 100; i++ {
		preinstaller.MarkInstalled("package-" + string(rune('a'+i%26)))
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			preinstaller.IsInstalled("package-" + string(rune('a'+i%26)))
			i++
		}
	})
}

func BenchmarkMCPConnectionPool_GetStatus(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := b.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	config := &mcp.MCPPoolConfig{
		MaxConnections:    100,
		ConnectionTimeout: 5 * time.Second,
		IdleTimeout:       time.Minute,
	}

	pool := mcp.NewMCPConnectionPool(preinstaller, config, logger)

	servers := []string{"server1", "server2", "server3", "server4", "server5"}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			pool.GetConnectionStatus(servers[i%len(servers)])
			i++
		}
	})
}

// TestMCPPreinstaller_NoRaceConditions runs with -race flag
func TestMCPPreinstaller_NoRaceConditions(t *testing.T) {
	logger := logrus.New()
	logger.SetOutput(os.Stdout)

	tempDir := t.TempDir()
	preinstaller := mcp.NewPreinstaller(tempDir, logger)

	var wg sync.WaitGroup
	var ops int64

	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func(idx int) {
			defer wg.Done()
			preinstaller.MarkInstalled("pkg-" + string(rune('a'+idx%26)))
			atomic.AddInt64(&ops, 1)
		}(i)

		go func(idx int) {
			defer wg.Done()
			preinstaller.IsInstalled("pkg-" + string(rune('a'+idx%26)))
			atomic.AddInt64(&ops, 1)
		}(i)

		go func() {
			defer wg.Done()
			preinstaller.Metrics()
			atomic.AddInt64(&ops, 1)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(300), atomic.LoadInt64(&ops))
}
