// Package adapters provides MCP server adapters.
// This file implements the Docker MCP server adapter.
package adapters

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"
)

// DockerConfig configures the Docker adapter.
type DockerConfig struct {
	Host       string        `json:"host"`
	APIVersion string        `json:"api_version"`
	TLSVerify  bool          `json:"tls_verify"`
	CertPath   string        `json:"cert_path,omitempty"`
	Timeout    time.Duration `json:"timeout"`
}

// DefaultDockerConfig returns default configuration.
func DefaultDockerConfig() DockerConfig {
	return DockerConfig{
		Host:       "unix:///var/run/docker.sock",
		APIVersion: "1.43",
		TLSVerify:  false,
		Timeout:    60 * time.Second,
	}
}

// DockerAdapter implements the Docker MCP server.
type DockerAdapter struct {
	config DockerConfig
	client DockerClient
}

// DockerClient interface for Docker operations.
type DockerClient interface {
	ListContainers(ctx context.Context, all bool) ([]Container, error)
	GetContainer(ctx context.Context, id string) (*ContainerDetails, error)
	CreateContainer(ctx context.Context, config ContainerConfig) (string, error)
	StartContainer(ctx context.Context, id string) error
	StopContainer(ctx context.Context, id string, timeout int) error
	RestartContainer(ctx context.Context, id string, timeout int) error
	RemoveContainer(ctx context.Context, id string, force bool) error
	ContainerLogs(ctx context.Context, id string, tail int) (io.ReadCloser, error)
	ExecInContainer(ctx context.Context, id string, cmd []string) (string, error)
	ListImages(ctx context.Context) ([]Image, error)
	PullImage(ctx context.Context, ref string) error
	RemoveImage(ctx context.Context, id string, force bool) error
	BuildImage(ctx context.Context, dockerfile, tag string) error
	ListNetworks(ctx context.Context) ([]Network, error)
	CreateNetwork(ctx context.Context, name, driver string) (string, error)
	RemoveNetwork(ctx context.Context, id string) error
	ListVolumes(ctx context.Context) ([]Volume, error)
	CreateVolume(ctx context.Context, name string) (*Volume, error)
	RemoveVolume(ctx context.Context, name string, force bool) error
	SystemInfo(ctx context.Context) (*SystemInfo, error)
}

// Container represents a Docker container.
type Container struct {
	ID      string            `json:"id"`
	Names   []string          `json:"names"`
	Image   string            `json:"image"`
	ImageID string            `json:"imageId"`
	Command string            `json:"command"`
	Created time.Time         `json:"created"`
	State   string            `json:"state"`
	Status  string            `json:"status"`
	Ports   []Port            `json:"ports"`
	Labels  map[string]string `json:"labels"`
}

// ContainerDetails represents detailed container information.
type ContainerDetails struct {
	Container
	Config          ContainerConfig `json:"config"`
	NetworkSettings NetworkSettings `json:"networkSettings"`
	Mounts          []Mount         `json:"mounts"`
}

// ContainerConfig represents container configuration.
type ContainerConfig struct {
	Image        string              `json:"image"`
	Cmd          []string            `json:"cmd,omitempty"`
	Env          []string            `json:"env,omitempty"`
	WorkingDir   string              `json:"workingDir,omitempty"`
	ExposedPorts map[string]struct{} `json:"exposedPorts,omitempty"`
	Volumes      map[string]struct{} `json:"volumes,omitempty"`
	Labels       map[string]string   `json:"labels,omitempty"`
	Hostname     string              `json:"hostname,omitempty"`
	User         string              `json:"user,omitempty"`
}

// Port represents a container port.
type Port struct {
	IP          string `json:"ip,omitempty"`
	PrivatePort int    `json:"privatePort"`
	PublicPort  int    `json:"publicPort,omitempty"`
	Type        string `json:"type"`
}

// NetworkSettings represents container network settings.
type NetworkSettings struct {
	IPAddress string                      `json:"ipAddress"`
	Networks  map[string]EndpointSettings `json:"networks"`
}

// EndpointSettings represents network endpoint settings.
type EndpointSettings struct {
	IPAddress string `json:"ipAddress"`
	Gateway   string `json:"gateway"`
}

// Mount represents a container mount.
type Mount struct {
	Type        string `json:"type"`
	Source      string `json:"source"`
	Destination string `json:"destination"`
	Mode        string `json:"mode"`
}

// Image represents a Docker image.
type Image struct {
	ID          string    `json:"id"`
	RepoTags    []string  `json:"repoTags"`
	RepoDigests []string  `json:"repoDigests"`
	Created     time.Time `json:"created"`
	Size        int64     `json:"size"`
}

// Network represents a Docker network.
type Network struct {
	ID         string            `json:"id"`
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Scope      string            `json:"scope"`
	Internal   bool              `json:"internal"`
	Containers map[string]string `json:"containers"`
}

// Volume represents a Docker volume.
type Volume struct {
	Name       string            `json:"name"`
	Driver     string            `json:"driver"`
	Mountpoint string            `json:"mountpoint"`
	Labels     map[string]string `json:"labels"`
	Scope      string            `json:"scope"`
}

// SystemInfo represents Docker system information.
type SystemInfo struct {
	Containers        int    `json:"containers"`
	ContainersRunning int    `json:"containersRunning"`
	ContainersPaused  int    `json:"containersPaused"`
	ContainersStopped int    `json:"containersStopped"`
	Images            int    `json:"images"`
	ServerVersion     string `json:"serverVersion"`
	OperatingSystem   string `json:"operatingSystem"`
	Architecture      string `json:"architecture"`
	NCPU              int    `json:"ncpu"`
	MemTotal          int64  `json:"memTotal"`
}

// NewDockerAdapter creates a new Docker adapter.
func NewDockerAdapter(config DockerConfig, client DockerClient) *DockerAdapter {
	return &DockerAdapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *DockerAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "docker",
		Version:     "1.0.0",
		Description: "Docker container management including containers, images, networks, and volumes",
		Capabilities: []string{
			"containers",
			"images",
			"networks",
			"volumes",
			"exec",
		},
	}
}

// ListTools returns available tools.
func (a *DockerAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "docker_list_containers",
			Description: "List Docker containers",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"all": map[string]interface{}{
						"type":        "boolean",
						"description": "Include stopped containers",
						"default":     false,
					},
				},
			},
		},
		{
			Name:        "docker_get_container",
			Description: "Get container details",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Container ID or name",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "docker_create_container",
			Description: "Create a new container",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Image name",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Container name",
					},
					"cmd": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Command to run",
					},
					"env": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Environment variables (KEY=VALUE)",
					},
				},
				"required": []string{"image"},
			},
		},
		{
			Name:        "docker_start_container",
			Description: "Start a container",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Container ID or name",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "docker_stop_container",
			Description: "Stop a container",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Container ID or name",
					},
					"timeout": map[string]interface{}{
						"type":        "integer",
						"description": "Timeout in seconds",
						"default":     10,
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "docker_restart_container",
			Description: "Restart a container",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Container ID or name",
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "docker_remove_container",
			Description: "Remove a container",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Container ID or name",
					},
					"force": map[string]interface{}{
						"type":        "boolean",
						"description": "Force removal",
						"default":     false,
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "docker_container_logs",
			Description: "Get container logs",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Container ID or name",
					},
					"tail": map[string]interface{}{
						"type":        "integer",
						"description": "Number of lines",
						"default":     100,
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "docker_exec",
			Description: "Execute a command in a container",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Container ID or name",
					},
					"cmd": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Command to execute",
					},
				},
				"required": []string{"id", "cmd"},
			},
		},
		{
			Name:        "docker_list_images",
			Description: "List Docker images",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "docker_pull_image",
			Description: "Pull an image from registry",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"image": map[string]interface{}{
						"type":        "string",
						"description": "Image reference",
					},
				},
				"required": []string{"image"},
			},
		},
		{
			Name:        "docker_remove_image",
			Description: "Remove an image",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"id": map[string]interface{}{
						"type":        "string",
						"description": "Image ID or name",
					},
					"force": map[string]interface{}{
						"type":        "boolean",
						"description": "Force removal",
						"default":     false,
					},
				},
				"required": []string{"id"},
			},
		},
		{
			Name:        "docker_list_networks",
			Description: "List Docker networks",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "docker_list_volumes",
			Description: "List Docker volumes",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "docker_system_info",
			Description: "Get Docker system information",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
	}
}

// CallTool executes a tool.
func (a *DockerAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "docker_list_containers":
		return a.listContainers(ctx, args)
	case "docker_get_container":
		return a.getContainer(ctx, args)
	case "docker_create_container":
		return a.createContainer(ctx, args)
	case "docker_start_container":
		return a.startContainer(ctx, args)
	case "docker_stop_container":
		return a.stopContainer(ctx, args)
	case "docker_restart_container":
		return a.restartContainer(ctx, args)
	case "docker_remove_container":
		return a.removeContainer(ctx, args)
	case "docker_container_logs":
		return a.containerLogs(ctx, args)
	case "docker_exec":
		return a.execInContainer(ctx, args)
	case "docker_list_images":
		return a.listImages(ctx)
	case "docker_pull_image":
		return a.pullImage(ctx, args)
	case "docker_remove_image":
		return a.removeImage(ctx, args)
	case "docker_list_networks":
		return a.listNetworks(ctx)
	case "docker_list_volumes":
		return a.listVolumes(ctx)
	case "docker_system_info":
		return a.systemInfo(ctx)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *DockerAdapter) listContainers(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	all, _ := args["all"].(bool)

	containers, err := a.client.ListContainers(ctx, all)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d containers:\n\n", len(containers)))

	for _, c := range containers {
		icon := "🟢"
		if c.State != "running" {
			icon = "⚪"
		}
		name := ""
		if len(c.Names) > 0 {
			name = c.Names[0]
		}
		sb.WriteString(fmt.Sprintf("%s %s (%s)\n", icon, name, c.ID[:12]))
		sb.WriteString(fmt.Sprintf("   Image: %s, Status: %s\n", c.Image, c.Status))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DockerAdapter) getContainer(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)

	container, err := a.client.GetContainer(ctx, id)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Container: %s\n", container.ID[:12]))
	sb.WriteString(fmt.Sprintf("Image: %s\n", container.Image))
	sb.WriteString(fmt.Sprintf("State: %s\n", container.State))
	sb.WriteString(fmt.Sprintf("Status: %s\n", container.Status))
	sb.WriteString(fmt.Sprintf("Created: %s\n", container.Created.Format(time.RFC3339)))

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DockerAdapter) createContainer(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	image, _ := args["image"].(string)
	cmdRaw, _ := args["cmd"].([]interface{})
	envRaw, _ := args["env"].([]interface{})

	var cmd []string
	for _, c := range cmdRaw {
		if s, ok := c.(string); ok {
			cmd = append(cmd, s)
		}
	}

	var env []string
	for _, e := range envRaw {
		if s, ok := e.(string); ok {
			env = append(env, s)
		}
	}

	config := ContainerConfig{
		Image: image,
		Cmd:   cmd,
		Env:   env,
	}

	id, err := a.client.CreateContainer(ctx, config)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Created container: %s", id[:12])}},
	}, nil
}

func (a *DockerAdapter) startContainer(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)

	err := a.client.StartContainer(ctx, id)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Started container: %s", id)}},
	}, nil
}

func (a *DockerAdapter) stopContainer(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)
	timeout := getIntArg(args, "timeout", 10)

	err := a.client.StopContainer(ctx, id, timeout)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Stopped container: %s", id)}},
	}, nil
}

func (a *DockerAdapter) restartContainer(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)

	err := a.client.RestartContainer(ctx, id, 10)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Restarted container: %s", id)}},
	}, nil
}

func (a *DockerAdapter) removeContainer(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)
	force, _ := args["force"].(bool)

	err := a.client.RemoveContainer(ctx, id, force)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Removed container: %s", id)}},
	}, nil
}

func (a *DockerAdapter) containerLogs(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)
	tail := getIntArg(args, "tail", 100)

	reader, err := a.client.ContainerLogs(ctx, id, tail)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}
	defer reader.Close()

	logs, err := io.ReadAll(reader)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: string(logs)}},
	}, nil
}

func (a *DockerAdapter) execInContainer(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)
	cmdRaw, _ := args["cmd"].([]interface{})

	var cmd []string
	for _, c := range cmdRaw {
		if s, ok := c.(string); ok {
			cmd = append(cmd, s)
		}
	}

	output, err := a.client.ExecInContainer(ctx, id, cmd)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: output}},
	}, nil
}

func (a *DockerAdapter) listImages(ctx context.Context) (*ToolResult, error) {
	images, err := a.client.ListImages(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d images:\n\n", len(images)))

	for _, img := range images {
		tags := "none"
		if len(img.RepoTags) > 0 {
			tags = strings.Join(img.RepoTags, ", ")
		}
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", tags, img.ID[:12]))
		sb.WriteString(fmt.Sprintf("  Size: %.2f MB\n", float64(img.Size)/1024/1024))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DockerAdapter) pullImage(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	image, _ := args["image"].(string)

	err := a.client.PullImage(ctx, image)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Pulled image: %s", image)}},
	}, nil
}

func (a *DockerAdapter) removeImage(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	id, _ := args["id"].(string)
	force, _ := args["force"].(bool)

	err := a.client.RemoveImage(ctx, id, force)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Removed image: %s", id)}},
	}, nil
}

func (a *DockerAdapter) listNetworks(ctx context.Context) (*ToolResult, error) {
	networks, err := a.client.ListNetworks(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d networks:\n\n", len(networks)))

	for _, n := range networks {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", n.Name, n.ID[:12]))
		sb.WriteString(fmt.Sprintf("  Driver: %s, Scope: %s\n", n.Driver, n.Scope))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DockerAdapter) listVolumes(ctx context.Context) (*ToolResult, error) {
	volumes, err := a.client.ListVolumes(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d volumes:\n\n", len(volumes)))

	for _, v := range volumes {
		sb.WriteString(fmt.Sprintf("- %s\n", v.Name))
		sb.WriteString(fmt.Sprintf("  Driver: %s, Mountpoint: %s\n", v.Driver, v.Mountpoint))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *DockerAdapter) systemInfo(ctx context.Context) (*ToolResult, error) {
	info, err := a.client.SystemInfo(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString("Docker System Info:\n\n")
	sb.WriteString(fmt.Sprintf("Server Version: %s\n", info.ServerVersion))
	sb.WriteString(fmt.Sprintf("OS: %s (%s)\n", info.OperatingSystem, info.Architecture))
	sb.WriteString(fmt.Sprintf("CPUs: %d\n", info.NCPU))
	sb.WriteString(fmt.Sprintf("Memory: %.2f GB\n", float64(info.MemTotal)/1024/1024/1024))
	sb.WriteString(fmt.Sprintf("Containers: %d (%d running, %d stopped)\n",
		info.Containers, info.ContainersRunning, info.ContainersStopped))
	sb.WriteString(fmt.Sprintf("Images: %d\n", info.Images))

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}
