package containers

import (
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"digital.vasic.containers/pkg/compose"
	"digital.vasic.containers/pkg/distribution"
	"digital.vasic.containers/pkg/health"
	"digital.vasic.containers/pkg/logging"
	"digital.vasic.containers/pkg/remote"
	"digital.vasic.containers/pkg/runtime"
	"digital.vasic.containers/pkg/scheduler"
)

// mockRuntime implements runtime.ContainerRuntime for testing.
type mockRuntime struct {
	name string
}

func (m *mockRuntime) Name() string { return m.name }

func (m *mockRuntime) Version(
	ctx context.Context,
) (string, error) {
	return "1.0.0", nil
}

func (m *mockRuntime) IsAvailable(ctx context.Context) bool {
	return true
}

func (m *mockRuntime) Start(
	ctx context.Context, id string, opts ...runtime.StartOption,
) error {
	return nil
}

func (m *mockRuntime) Stop(
	ctx context.Context, id string, opts ...runtime.StopOption,
) error {
	return nil
}

func (m *mockRuntime) Remove(
	ctx context.Context, id string, opts ...runtime.RemoveOption,
) error {
	return nil
}

func (m *mockRuntime) Status(
	ctx context.Context, id string,
) (*runtime.ContainerStatus, error) {
	return &runtime.ContainerStatus{
		Name:  id,
		State: runtime.StateRunning,
	}, nil
}

func (m *mockRuntime) List(
	ctx context.Context, filter runtime.ListFilter,
) ([]runtime.ContainerInfo, error) {
	return nil, nil
}

func (m *mockRuntime) Stats(
	ctx context.Context, id string,
) (*runtime.ContainerStats, error) {
	return &runtime.ContainerStats{}, nil
}

func (m *mockRuntime) Exec(
	ctx context.Context, id string, cmd []string,
) (*runtime.ExecResult, error) {
	return &runtime.ExecResult{ExitCode: 0}, nil
}

func (m *mockRuntime) Logs(
	ctx context.Context, id string, opts ...runtime.LogOption,
) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

// mockOrchestrator implements compose.ComposeOrchestrator.
type mockOrchestrator struct {
	upCalled   bool
	downCalled bool
	lastFile   string
}

func (m *mockOrchestrator) Up(
	ctx context.Context, project compose.ComposeProject,
	opts ...compose.UpOption,
) error {
	m.upCalled = true
	m.lastFile = project.File
	return nil
}

func (m *mockOrchestrator) Down(
	ctx context.Context, project compose.ComposeProject,
	opts ...compose.DownOption,
) error {
	m.downCalled = true
	m.lastFile = project.File
	return nil
}

func (m *mockOrchestrator) Status(
	ctx context.Context, project compose.ComposeProject,
) ([]compose.ServiceStatus, error) {
	return nil, nil
}

func (m *mockOrchestrator) Logs(
	ctx context.Context, project compose.ComposeProject,
	service string,
) (io.ReadCloser, error) {
	return io.NopCloser(strings.NewReader("")), nil
}

// mockDistributor implements distribution.Distributor.
type mockDistributor struct {
	distributeCalled   bool
	undistributeCalled bool
	containers         []distribution.DistributedContainer
}

func (m *mockDistributor) Distribute(
	ctx context.Context,
	reqs []scheduler.ContainerRequirements,
) (*distribution.DistributionSummary, error) {
	m.distributeCalled = true
	return &distribution.DistributionSummary{
		TotalContainers: len(reqs),
		LocalContainers: len(reqs),
	}, nil
}

func (m *mockDistributor) Undistribute(
	ctx context.Context,
) error {
	m.undistributeCalled = true
	return nil
}

func (m *mockDistributor) Status(
	ctx context.Context,
) []distribution.DistributedContainer {
	return m.containers
}

func (m *mockDistributor) HealthCheckAll(
	ctx context.Context,
) map[string]error {
	return nil
}

func (m *mockDistributor) Rebalance(
	ctx context.Context,
) (*distribution.DistributionSummary, error) {
	return &distribution.DistributionSummary{}, nil
}

func (m *mockDistributor) HostStatus(
	ctx context.Context, hostName string,
) (*remote.HostResources, error) {
	return &remote.HostResources{Host: hostName}, nil
}

// mockHostManager implements remote.HostManager.
type mockHostManager struct {
	hosts map[string]remote.RemoteHost
}

func (m *mockHostManager) AddHost(h remote.RemoteHost) error {
	m.hosts[h.Name] = h
	return nil
}

func (m *mockHostManager) RemoveHost(name string) error {
	delete(m.hosts, name)
	return nil
}

func (m *mockHostManager) GetHost(
	name string,
) (*remote.RemoteHost, error) {
	h, ok := m.hosts[name]
	if !ok {
		return nil, nil
	}
	return &h, nil
}

func (m *mockHostManager) ListHosts() []remote.RemoteHost {
	hosts := make([]remote.RemoteHost, 0, len(m.hosts))
	for _, h := range m.hosts {
		hosts = append(hosts, h)
	}
	return hosts
}

func (m *mockHostManager) ProbeHost(
	ctx context.Context, name string,
) (*remote.HostResources, error) {
	return &remote.HostResources{Host: name}, nil
}

func (m *mockHostManager) ProbeAll(
	ctx context.Context,
) map[string]*remote.HostResources {
	return nil
}

func (m *mockHostManager) HostState(
	name string,
) remote.HostState {
	return remote.HostOnline
}

func TestNewAdapter(t *testing.T) {
	adapter, err := NewAdapter(
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)
	assert.NotNil(t, adapter)
}

func TestNewAdapter_WithProjectDir(t *testing.T) {
	adapter, err := NewAdapter(
		WithProjectDir("/tmp/test"),
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/test", adapter.projectDir)
}

func TestAdapter_DetectRuntime_WithExisting(t *testing.T) {
	rt := &mockRuntime{name: "docker"}
	adapter, err := NewAdapter(
		WithRuntime(rt),
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	name, err := adapter.DetectRuntime(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "docker", name)
}

func TestAdapter_RuntimeAvailable(t *testing.T) {
	rt := &mockRuntime{name: "docker"}
	adapter, err := NewAdapter(
		WithRuntime(rt),
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	assert.True(t, adapter.RuntimeAvailable(
		context.Background(),
	))
}

func TestAdapter_ComposeUp(t *testing.T) {
	orch := &mockOrchestrator{}
	adapter, err := NewAdapter(
		WithOrchestrator(orch),
		WithProjectDir("/tmp/project"),
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	err = adapter.ComposeUp(
		context.Background(),
		"docker-compose.yml", "default",
	)
	require.NoError(t, err)
	assert.True(t, orch.upCalled)
	assert.Contains(t, orch.lastFile, "docker-compose.yml")
}

func TestAdapter_ComposeDown(t *testing.T) {
	orch := &mockOrchestrator{}
	adapter, err := NewAdapter(
		WithOrchestrator(orch),
		WithProjectDir("/tmp/project"),
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	err = adapter.ComposeDown(
		context.Background(),
		"docker-compose.yml", "",
	)
	require.NoError(t, err)
	assert.True(t, orch.downCalled)
}

func TestAdapter_ComposeUp_NoOrchestrator(t *testing.T) {
	adapter, err := NewAdapter(
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	err = adapter.ComposeUp(
		context.Background(),
		"docker-compose.yml", "",
	)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not available")
}

func TestAdapter_ComposeStatus_NoOrchestrator(t *testing.T) {
	adapter, err := NewAdapter(
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	_, err = adapter.ComposeStatus(
		context.Background(),
		"docker-compose.yml",
	)
	assert.Error(t, err)
}

func TestAdapter_HealthCheck_NoChecker(t *testing.T) {
	adapter, err := NewAdapter(
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	_, err = adapter.HealthCheck(
		context.Background(),
		"test", "localhost", "8080",
		"/health", "http", 5*time.Second,
	)
	assert.Error(t, err)
}

func TestAdapter_HealthCheckHTTP_InvalidURL(t *testing.T) {
	adapter, err := NewAdapter(
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	err = adapter.HealthCheckHTTP("http://localhost:99999/invalid")
	assert.Error(t, err)
}

func TestAdapter_HealthCheckTCP_InvalidPort(t *testing.T) {
	adapter, err := NewAdapter(
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	ok := adapter.HealthCheckTCP("localhost", 1)
	assert.False(t, ok)
}

func TestAdapter_Distribute(t *testing.T) {
	dist := &mockDistributor{}
	adapter, err := NewAdapter(
		WithDistributor(dist),
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	reqs := []scheduler.ContainerRequirements{
		{Name: "app-1", Image: "nginx"},
	}
	summary, err := adapter.Distribute(
		context.Background(), reqs,
	)
	require.NoError(t, err)
	assert.True(t, dist.distributeCalled)
	assert.Equal(t, 1, summary.TotalContainers)
}

func TestAdapter_Distribute_NoDistributor(t *testing.T) {
	adapter, err := NewAdapter(
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	_, err = adapter.Distribute(
		context.Background(),
		[]scheduler.ContainerRequirements{{Name: "app"}},
	)
	assert.Error(t, err)
}

func TestAdapter_Undistribute(t *testing.T) {
	dist := &mockDistributor{}
	adapter, err := NewAdapter(
		WithDistributor(dist),
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	err = adapter.Undistribute(context.Background())
	require.NoError(t, err)
	assert.True(t, dist.undistributeCalled)
}

func TestAdapter_Undistribute_NoDistributor(t *testing.T) {
	adapter, err := NewAdapter(
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	err = adapter.Undistribute(context.Background())
	assert.NoError(t, err)
}

func TestAdapter_DistributionStatus(t *testing.T) {
	containers := []distribution.DistributedContainer{
		{HostName: "local", State: distribution.StateRunning},
	}
	dist := &mockDistributor{containers: containers}
	adapter, err := NewAdapter(
		WithDistributor(dist),
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	status := adapter.DistributionStatus(context.Background())
	assert.Len(t, status, 1)
}

func TestAdapter_RemoteEnabled(t *testing.T) {
	adapter, _ := NewAdapter(WithLogger(logging.NopLogger{}))
	assert.False(t, adapter.RemoteEnabled())

	adapter.distributor = &mockDistributor{}
	adapter.hostManager = &mockHostManager{
		hosts: map[string]remote.RemoteHost{},
	}
	assert.True(t, adapter.RemoteEnabled())
}

func TestAdapter_ListHosts(t *testing.T) {
	hm := &mockHostManager{
		hosts: map[string]remote.RemoteHost{
			"h1": {Name: "h1", Address: "10.0.0.1"},
		},
	}
	adapter, _ := NewAdapter(
		WithHostManager(hm),
		WithLogger(logging.NopLogger{}),
	)

	hosts := adapter.ListHosts()
	assert.Len(t, hosts, 1)
}

func TestAdapter_ListHosts_NoManager(t *testing.T) {
	adapter, _ := NewAdapter(WithLogger(logging.NopLogger{}))
	hosts := adapter.ListHosts()
	assert.Nil(t, hosts)
}

func TestAdapter_ProbeHost(t *testing.T) {
	hm := &mockHostManager{
		hosts: map[string]remote.RemoteHost{},
	}
	adapter, _ := NewAdapter(
		WithHostManager(hm),
		WithLogger(logging.NopLogger{}),
	)

	res, err := adapter.ProbeHost(
		context.Background(), "h1",
	)
	require.NoError(t, err)
	assert.Equal(t, "h1", res.Host)
}

func TestAdapter_ProbeHost_NoManager(t *testing.T) {
	adapter, _ := NewAdapter(WithLogger(logging.NopLogger{}))
	_, err := adapter.ProbeHost(
		context.Background(), "h1",
	)
	assert.Error(t, err)
}

func TestAdapter_Shutdown(t *testing.T) {
	dist := &mockDistributor{}
	adapter, _ := NewAdapter(
		WithDistributor(dist),
		WithLogger(logging.NopLogger{}),
	)

	err := adapter.Shutdown(context.Background())
	require.NoError(t, err)
	assert.True(t, dist.undistributeCalled)
}

func TestAdapter_Runtime(t *testing.T) {
	rt := &mockRuntime{name: "podman"}
	adapter, _ := NewAdapter(
		WithRuntime(rt),
		WithLogger(logging.NopLogger{}),
	)

	assert.Equal(t, "podman", adapter.Runtime().Name())
}

func TestAdapter_Orchestrator(t *testing.T) {
	orch := &mockOrchestrator{}
	adapter, _ := NewAdapter(
		WithOrchestrator(orch),
		WithLogger(logging.NopLogger{}),
	)

	assert.NotNil(t, adapter.Orchestrator())
}

func TestAdapter_ToEndpoint(t *testing.T) {
	adapter, _ := NewAdapter(WithLogger(logging.NopLogger{}))
	ep := adapter.ToEndpoint(
		"postgres", "localhost", "5432",
		"/health", "tcp",
		"docker-compose.yml", "postgres", "default",
		true, true, false,
	)
	assert.Equal(t, "localhost", ep.Host)
	assert.Equal(t, "5432", ep.Port)
	assert.True(t, ep.Required)
	assert.False(t, ep.Remote)
}

func TestAdapter_HealthCheck_WithChecker(t *testing.T) {
	checker := health.NewDefaultChecker()
	adapter, err := NewAdapter(
		WithHealthChecker(checker),
		WithLogger(logging.NopLogger{}),
	)
	require.NoError(t, err)

	// TCP check to a port that should be closed.
	result, err := adapter.HealthCheck(
		context.Background(),
		"test", "localhost", "1",
		"", "tcp", 1*time.Second,
	)
	require.NoError(t, err)
	assert.False(t, result.Healthy)
}

func TestLogrusAdapter(t *testing.T) {
	l := &logrusAdapter{}
	// Just verify no panic.
	l.Debug("debug %s", "test")
	l.Info("info %s", "test")
	l.Warn("warn %s", "test")
	l.Error("error %s", "test")
}
