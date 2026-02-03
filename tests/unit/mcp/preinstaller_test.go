package mcp

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	mcp "dev.helix.agent/internal/mcp"
)

func TestMCPPreinstaller_NewPreinstaller(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()

	config := mcp.PreinstallerConfig{
		InstallDir: tempDir,
		Logger:     logger,
	}

	preinstaller, err := mcp.NewPreinstaller(config)
	require.NoError(t, err)
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
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	config := mcp.PreinstallerConfig{
		InstallDir: tempDir,
		Logger:     logger,
	}
	preinstaller, err := mcp.NewPreinstaller(config)
	require.NoError(t, err)

	// Package not installed yet
	installed := preinstaller.IsInstalled("nonexistent-package")
	assert.False(t, installed)
}

func TestMCPPreinstaller_GetStatus(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	config := mcp.PreinstallerConfig{
		InstallDir: tempDir,
		Logger:     logger,
	}
	preinstaller, err := mcp.NewPreinstaller(config)
	require.NoError(t, err)

	// Get status of a standard package
	status, err := preinstaller.GetStatus("filesystem")
	require.NoError(t, err)
	assert.Equal(t, mcp.StatusPending, status.Status)
}

func TestMCPPreinstaller_GetAllStatuses(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	config := mcp.PreinstallerConfig{
		InstallDir: tempDir,
		Logger:     logger,
	}
	preinstaller, err := mcp.NewPreinstaller(config)
	require.NoError(t, err)

	statuses := preinstaller.GetAllStatuses()
	require.NotEmpty(t, statuses)

	// Check that standard packages are in statuses
	assert.Contains(t, statuses, "filesystem")
	assert.Contains(t, statuses, "github")
	assert.Contains(t, statuses, "memory")
}

func TestMCPPreinstaller_WaitForPackage_Timeout(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	config := mcp.PreinstallerConfig{
		InstallDir: tempDir,
		Logger:     logger,
	}
	preinstaller, err := mcp.NewPreinstaller(config)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err = preinstaller.WaitForPackage(ctx, "filesystem")
	assert.Error(t, err)
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestMCPPreinstaller_GetInstallDir(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	config := mcp.PreinstallerConfig{
		InstallDir: tempDir,
		Logger:     logger,
	}
	preinstaller, err := mcp.NewPreinstaller(config)
	require.NoError(t, err)

	installDir := preinstaller.GetInstallDir()
	assert.Equal(t, tempDir, installDir)
}

func TestMCPPreinstaller_IsNodeAvailable(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	config := mcp.PreinstallerConfig{
		InstallDir: tempDir,
		Logger:     logger,
	}
	preinstaller, err := mcp.NewPreinstaller(config)
	require.NoError(t, err)

	// This will return true or false depending on whether node is installed
	_ = preinstaller.IsNodeAvailable()
}

func TestMCPConnectionPool_NewPool(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)
	require.NotNil(t, pool)
	defer pool.Close()
}

func TestMCPConnectionPool_RegisterServer(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)
	defer pool.Close()

	serverConfig := mcp.MCPServerConfig{
		Name:    "test-server",
		Type:    mcp.MCPServerTypeLocal,
		Command: []string{"node", "server.js"},
		Enabled: true,
	}

	err := pool.RegisterServer(serverConfig)
	require.NoError(t, err)

	// Registering same server again should fail
	err = pool.RegisterServer(serverConfig)
	assert.Error(t, err)
}

func TestMCPConnectionPool_ListServers(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)
	defer pool.Close()

	pool.RegisterServer(mcp.MCPServerConfig{
		Name:    "server1",
		Type:    mcp.MCPServerTypeLocal,
		Command: []string{"node", "server1.js"},
	})
	pool.RegisterServer(mcp.MCPServerConfig{
		Name:    "server2",
		Type:    mcp.MCPServerTypeLocal,
		Command: []string{"node", "server2.js"},
	})

	servers := pool.ListServers()
	assert.Len(t, servers, 2)
	assert.Contains(t, servers, "server1")
	assert.Contains(t, servers, "server2")
}

func TestMCPConnectionPool_GetServerStatus(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)
	defer pool.Close()

	pool.RegisterServer(mcp.MCPServerConfig{
		Name:    "test-server",
		Type:    mcp.MCPServerTypeLocal,
		Command: []string{"node", "server.js"},
	})

	status, err := pool.GetServerStatus("test-server")
	require.NoError(t, err)
	assert.Equal(t, mcp.StatusConnectionPending, status)

	// Non-existent server should error
	_, err = pool.GetServerStatus("nonexistent")
	assert.Error(t, err)
}

func TestMCPConnectionPool_GetMetrics(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)
	defer pool.Close()

	metrics := pool.GetMetrics()
	require.NotNil(t, metrics)
	assert.Equal(t, int64(0), metrics.ActiveConnections)
	assert.Equal(t, int64(0), metrics.TotalConnections)
}

func TestMCPConnectionPool_Close(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)

	pool.RegisterServer(mcp.MCPServerConfig{
		Name:    "test-server",
		Type:    mcp.MCPServerTypeLocal,
		Command: []string{"node", "server.js"},
	})

	// Close should not panic
	err := pool.Close()
	assert.NoError(t, err)

	// Registering after close should fail
	err = pool.RegisterServer(mcp.MCPServerConfig{
		Name: "new-server",
	})
	assert.Error(t, err)
}

func TestMCPConnectionPool_ConcurrentAccess(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)
	defer pool.Close()

	// Register some servers
	for i := 0; i < 5; i++ {
		pool.RegisterServer(mcp.MCPServerConfig{
			Name:    "server" + string(rune('0'+i)),
			Type:    mcp.MCPServerTypeLocal,
			Command: []string{"node", "server.js"},
		})
	}

	var wg sync.WaitGroup
	servers := pool.ListServers()

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			server := servers[idx%len(servers)]
			pool.GetServerStatus(server)
			pool.GetMetrics()
		}(i)
	}

	wg.Wait()
}

func TestMCPConnectionPool_HealthCheck(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)
	defer pool.Close()

	pool.RegisterServer(mcp.MCPServerConfig{
		Name:    "test-server",
		Type:    mcp.MCPServerTypeLocal,
		Command: []string{"node", "server.js"},
	})

	ctx := context.Background()
	results := pool.HealthCheck(ctx)

	assert.Contains(t, results, "test-server")
	// Not connected, so health check should be false
	assert.False(t, results["test-server"])
}

func TestConnectionStatus_Constants(t *testing.T) {
	// Test that connection status constants exist and have expected string values
	assert.Equal(t, mcp.ConnectionStatus("pending"), mcp.StatusConnectionPending)
	assert.Equal(t, mcp.ConnectionStatus("connecting"), mcp.StatusConnectionConnecting)
	assert.Equal(t, mcp.ConnectionStatus("connected"), mcp.StatusConnectionConnected)
	assert.Equal(t, mcp.ConnectionStatus("failed"), mcp.StatusConnectionFailed)
	assert.Equal(t, mcp.ConnectionStatus("closed"), mcp.StatusConnectionClosed)
}

func TestMCPPoolConfig_Defaults(t *testing.T) {
	config := mcp.DefaultPoolConfig()

	assert.True(t, config.MaxConnections > 0)
	assert.True(t, config.ConnectionTimeout > 0)
	assert.True(t, config.IdleTimeout > 0)
}

func TestMCPServerType_Constants(t *testing.T) {
	assert.Equal(t, mcp.MCPServerType("local"), mcp.MCPServerTypeLocal)
	assert.Equal(t, mcp.MCPServerType("remote"), mcp.MCPServerTypeRemote)
}

func TestMCPPreinstaller_PreInstallAll_NoNPM(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	tempDir := t.TempDir()
	config := mcp.PreinstallerConfig{
		InstallDir: tempDir,
		Logger:     logger,
		Packages: []mcp.MCPPackage{
			{Name: "test", NPM: "test-package"},
		},
	}

	preinstaller, err := mcp.NewPreinstaller(config)
	require.NoError(t, err)

	// If npm is not available, PreInstallAll should return without error
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// This test depends on whether npm is available
	// If npm is not available, it should silently skip
	_ = preinstaller.PreInstallAll(ctx)
}

func BenchmarkMCPConnectionPool_GetStatus(b *testing.B) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)
	defer pool.Close()

	servers := []string{"server1", "server2", "server3", "server4", "server5"}
	for _, name := range servers {
		pool.RegisterServer(mcp.MCPServerConfig{
			Name:    name,
			Type:    mcp.MCPServerTypeLocal,
			Command: []string{"node", "server.js"},
		})
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			pool.GetServerStatus(servers[i%len(servers)])
			i++
		}
	})
}

// TestMCPConnectionPool_NoRaceConditions runs with -race flag
func TestMCPConnectionPool_NoRaceConditions(t *testing.T) {
	logger := logrus.New()
	logger.SetLevel(logrus.ErrorLevel)

	config := mcp.DefaultPoolConfig()
	pool := mcp.NewConnectionPool(nil, config, logger)
	defer pool.Close()

	// Register servers
	for i := 0; i < 10; i++ {
		pool.RegisterServer(mcp.MCPServerConfig{
			Name:    "server" + string(rune('0'+i)),
			Type:    mcp.MCPServerTypeLocal,
			Command: []string{"node", "server.js"},
		})
	}

	var wg sync.WaitGroup
	var ops int64

	servers := pool.ListServers()

	for i := 0; i < 100; i++ {
		wg.Add(3)

		go func(idx int) {
			defer wg.Done()
			pool.GetServerStatus(servers[idx%len(servers)])
			atomic.AddInt64(&ops, 1)
		}(i)

		go func() {
			defer wg.Done()
			pool.GetMetrics()
			atomic.AddInt64(&ops, 1)
		}()

		go func() {
			defer wg.Done()
			pool.ListServers()
			atomic.AddInt64(&ops, 1)
		}()
	}

	wg.Wait()
	assert.Equal(t, int64(300), atomic.LoadInt64(&ops))
}
