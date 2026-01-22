// Package adapters provides MCP server adapters.
// This file implements the Kubernetes MCP server adapter.
package adapters

import (
	"context"
	"fmt"
	"strings"
	"time"
)

// KubernetesConfig configures the Kubernetes adapter.
type KubernetesConfig struct {
	Kubeconfig string        `json:"kubeconfig,omitempty"`
	Context    string        `json:"context,omitempty"`
	Namespace  string        `json:"namespace"`
	Timeout    time.Duration `json:"timeout"`
	InCluster  bool          `json:"in_cluster"`
}

// DefaultKubernetesConfig returns default configuration.
func DefaultKubernetesConfig() KubernetesConfig {
	return KubernetesConfig{
		Namespace: "default",
		Timeout:   60 * time.Second,
		InCluster: false,
	}
}

// KubernetesAdapter implements the Kubernetes MCP server.
type KubernetesAdapter struct {
	config KubernetesConfig
	client KubernetesClient
}

// KubernetesClient interface for Kubernetes operations.
type KubernetesClient interface {
	// Pods
	ListPods(ctx context.Context, namespace string, labelSelector string) ([]Pod, error)
	GetPod(ctx context.Context, namespace, name string) (*Pod, error)
	DeletePod(ctx context.Context, namespace, name string) error
	GetPodLogs(ctx context.Context, namespace, name, container string, tailLines int) (string, error)
	ExecInPod(ctx context.Context, namespace, name, container string, cmd []string) (string, error)
	// Deployments
	ListDeployments(ctx context.Context, namespace string) ([]Deployment, error)
	GetDeployment(ctx context.Context, namespace, name string) (*Deployment, error)
	ScaleDeployment(ctx context.Context, namespace, name string, replicas int) error
	RestartDeployment(ctx context.Context, namespace, name string) error
	// Services
	ListServices(ctx context.Context, namespace string) ([]Service, error)
	GetService(ctx context.Context, namespace, name string) (*Service, error)
	// ConfigMaps & Secrets
	ListConfigMaps(ctx context.Context, namespace string) ([]ConfigMap, error)
	GetConfigMap(ctx context.Context, namespace, name string) (*ConfigMap, error)
	ListSecrets(ctx context.Context, namespace string) ([]Secret, error)
	// Namespaces
	ListNamespaces(ctx context.Context) ([]Namespace, error)
	// Nodes
	ListNodes(ctx context.Context) ([]Node, error)
	// Events
	GetEvents(ctx context.Context, namespace string, fieldSelector string) ([]Event, error)
	// Apply
	Apply(ctx context.Context, yaml string, namespace string) error
	Delete(ctx context.Context, kind, namespace, name string) error
}

// Pod represents a Kubernetes pod.
type Pod struct {
	Name         string            `json:"name"`
	Namespace    string            `json:"namespace"`
	Status       string            `json:"status"`
	Phase        string            `json:"phase"`
	IP           string            `json:"ip"`
	Node         string            `json:"node"`
	Containers   []ContainerStatus `json:"containers"`
	Labels       map[string]string `json:"labels"`
	CreatedAt    time.Time         `json:"createdAt"`
	RestartCount int               `json:"restartCount"`
}

// ContainerStatus represents container status.
type ContainerStatus struct {
	Name         string `json:"name"`
	Ready        bool   `json:"ready"`
	RestartCount int    `json:"restartCount"`
	Image        string `json:"image"`
	State        string `json:"state"`
}

// Deployment represents a Kubernetes deployment.
type Deployment struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int               `json:"replicas"`
	ReadyReplicas     int               `json:"readyReplicas"`
	AvailableReplicas int               `json:"availableReplicas"`
	Labels            map[string]string `json:"labels"`
	Image             string            `json:"image"`
	CreatedAt         time.Time         `json:"createdAt"`
}

// Service represents a Kubernetes service.
type Service struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Type       string            `json:"type"`
	ClusterIP  string            `json:"clusterIP"`
	ExternalIP string            `json:"externalIP,omitempty"`
	Ports      []ServicePort     `json:"ports"`
	Labels     map[string]string `json:"labels"`
}

// ServicePort represents a service port.
type ServicePort struct {
	Name       string `json:"name"`
	Port       int    `json:"port"`
	TargetPort int    `json:"targetPort"`
	Protocol   string `json:"protocol"`
	NodePort   int    `json:"nodePort,omitempty"`
}

// ConfigMap represents a Kubernetes ConfigMap.
type ConfigMap struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Data      map[string]string `json:"data"`
	CreatedAt time.Time         `json:"createdAt"`
}

// Secret represents a Kubernetes Secret.
type Secret struct {
	Name      string    `json:"name"`
	Namespace string    `json:"namespace"`
	Type      string    `json:"type"`
	Keys      []string  `json:"keys"`
	CreatedAt time.Time `json:"createdAt"`
}

// Namespace represents a Kubernetes namespace.
type Namespace struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Labels    map[string]string `json:"labels"`
	CreatedAt time.Time         `json:"createdAt"`
}

// Node represents a Kubernetes node.
type Node struct {
	Name           string            `json:"name"`
	Status         string            `json:"status"`
	Roles          []string          `json:"roles"`
	Version        string            `json:"version"`
	InternalIP     string            `json:"internalIP"`
	OS             string            `json:"os"`
	Architecture   string            `json:"architecture"`
	CPUCapacity    string            `json:"cpuCapacity"`
	MemoryCapacity string            `json:"memoryCapacity"`
	Labels         map[string]string `json:"labels"`
}

// Event represents a Kubernetes event.
type Event struct {
	Type           string    `json:"type"`
	Reason         string    `json:"reason"`
	Message        string    `json:"message"`
	Source         string    `json:"source"`
	InvolvedObject string    `json:"involvedObject"`
	Count          int       `json:"count"`
	FirstTimestamp time.Time `json:"firstTimestamp"`
	LastTimestamp  time.Time `json:"lastTimestamp"`
}

// NewKubernetesAdapter creates a new Kubernetes adapter.
func NewKubernetesAdapter(config KubernetesConfig, client KubernetesClient) *KubernetesAdapter {
	return &KubernetesAdapter{
		config: config,
		client: client,
	}
}

// GetServerInfo returns server information.
func (a *KubernetesAdapter) GetServerInfo() ServerInfo {
	return ServerInfo{
		Name:        "kubernetes",
		Version:     "1.0.0",
		Description: "Kubernetes cluster management for pods, deployments, services, and more",
		Capabilities: []string{
			"pods",
			"deployments",
			"services",
			"configmaps",
			"secrets",
			"namespaces",
			"nodes",
			"events",
		},
	}
}

// ListTools returns available tools.
func (a *KubernetesAdapter) ListTools() []ToolDefinition {
	return []ToolDefinition{
		{
			Name:        "k8s_list_pods",
			Description: "List pods in a namespace",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace (default: default)",
						"default":     "default",
					},
					"label_selector": map[string]interface{}{
						"type":        "string",
						"description": "Label selector (e.g., app=nginx)",
					},
				},
			},
		},
		{
			Name:        "k8s_get_pod",
			Description: "Get pod details",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Pod name",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "k8s_delete_pod",
			Description: "Delete a pod",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Pod name",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "k8s_pod_logs",
			Description: "Get pod logs",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Pod name",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
					"container": map[string]interface{}{
						"type":        "string",
						"description": "Container name (optional)",
					},
					"tail_lines": map[string]interface{}{
						"type":        "integer",
						"description": "Number of lines to return",
						"default":     100,
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "k8s_exec_in_pod",
			Description: "Execute command in pod",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Pod name",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
					"container": map[string]interface{}{
						"type":        "string",
						"description": "Container name (optional)",
					},
					"command": map[string]interface{}{
						"type":        "array",
						"items":       map[string]interface{}{"type": "string"},
						"description": "Command to execute",
					},
				},
				"required": []string{"name", "command"},
			},
		},
		{
			Name:        "k8s_list_deployments",
			Description: "List deployments",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
				},
			},
		},
		{
			Name:        "k8s_scale_deployment",
			Description: "Scale a deployment",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Deployment name",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
					"replicas": map[string]interface{}{
						"type":        "integer",
						"description": "Number of replicas",
					},
				},
				"required": []string{"name", "replicas"},
			},
		},
		{
			Name:        "k8s_restart_deployment",
			Description: "Restart a deployment",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Deployment name",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
				},
				"required": []string{"name"},
			},
		},
		{
			Name:        "k8s_list_services",
			Description: "List services",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
				},
			},
		},
		{
			Name:        "k8s_list_namespaces",
			Description: "List namespaces",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "k8s_list_nodes",
			Description: "List cluster nodes",
			InputSchema: map[string]interface{}{
				"type":       "object",
				"properties": map[string]interface{}{},
			},
		},
		{
			Name:        "k8s_get_events",
			Description: "Get events in namespace",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
					"field_selector": map[string]interface{}{
						"type":        "string",
						"description": "Field selector",
					},
				},
			},
		},
		{
			Name:        "k8s_apply",
			Description: "Apply a YAML manifest",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"yaml": map[string]interface{}{
						"type":        "string",
						"description": "YAML manifest",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
				},
				"required": []string{"yaml"},
			},
		},
		{
			Name:        "k8s_delete",
			Description: "Delete a resource",
			InputSchema: map[string]interface{}{
				"type": "object",
				"properties": map[string]interface{}{
					"kind": map[string]interface{}{
						"type":        "string",
						"description": "Resource kind (pod, deployment, service, etc.)",
					},
					"name": map[string]interface{}{
						"type":        "string",
						"description": "Resource name",
					},
					"namespace": map[string]interface{}{
						"type":        "string",
						"description": "Namespace",
						"default":     "default",
					},
				},
				"required": []string{"kind", "name"},
			},
		},
	}
}

// CallTool executes a tool.
func (a *KubernetesAdapter) CallTool(ctx context.Context, name string, args map[string]interface{}) (*ToolResult, error) {
	switch name {
	case "k8s_list_pods":
		return a.listPods(ctx, args)
	case "k8s_get_pod":
		return a.getPod(ctx, args)
	case "k8s_delete_pod":
		return a.deletePod(ctx, args)
	case "k8s_pod_logs":
		return a.getPodLogs(ctx, args)
	case "k8s_exec_in_pod":
		return a.execInPod(ctx, args)
	case "k8s_list_deployments":
		return a.listDeployments(ctx, args)
	case "k8s_scale_deployment":
		return a.scaleDeployment(ctx, args)
	case "k8s_restart_deployment":
		return a.restartDeployment(ctx, args)
	case "k8s_list_services":
		return a.listServices(ctx, args)
	case "k8s_list_namespaces":
		return a.listNamespaces(ctx)
	case "k8s_list_nodes":
		return a.listNodes(ctx)
	case "k8s_get_events":
		return a.getEvents(ctx, args)
	case "k8s_apply":
		return a.apply(ctx, args)
	case "k8s_delete":
		return a.deleteResource(ctx, args)
	default:
		return nil, fmt.Errorf("unknown tool: %s", name)
	}
}

func (a *KubernetesAdapter) getNamespace(args map[string]interface{}) string {
	ns, _ := args["namespace"].(string)
	if ns == "" {
		return a.config.Namespace
	}
	return ns
}

func (a *KubernetesAdapter) listPods(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	namespace := a.getNamespace(args)
	labelSelector, _ := args["label_selector"].(string)

	pods, err := a.client.ListPods(ctx, namespace, labelSelector)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Pods in namespace '%s' (%d):\n\n", namespace, len(pods)))

	for _, p := range pods {
		icon := "🟢"
		if p.Phase != "Running" {
			icon = "⚪"
		}
		sb.WriteString(fmt.Sprintf("%s %s\n", icon, p.Name))
		sb.WriteString(fmt.Sprintf("   Status: %s, Restarts: %d, Node: %s\n", p.Phase, p.RestartCount, p.Node))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *KubernetesAdapter) getPod(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	namespace := a.getNamespace(args)

	pod, err := a.client.GetPod(ctx, namespace, name)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Pod: %s/%s\n", pod.Namespace, pod.Name))
	sb.WriteString(fmt.Sprintf("Status: %s\n", pod.Phase))
	sb.WriteString(fmt.Sprintf("IP: %s\n", pod.IP))
	sb.WriteString(fmt.Sprintf("Node: %s\n", pod.Node))
	sb.WriteString(fmt.Sprintf("Created: %s\n", pod.CreatedAt.Format(time.RFC3339)))
	sb.WriteString("\nContainers:\n")
	for _, c := range pod.Containers {
		ready := "✗"
		if c.Ready {
			ready = "✓"
		}
		sb.WriteString(fmt.Sprintf("  - %s [%s] %s (restarts: %d)\n", c.Name, ready, c.State, c.RestartCount))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *KubernetesAdapter) deletePod(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	namespace := a.getNamespace(args)

	err := a.client.DeletePod(ctx, namespace, name)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Deleted pod %s/%s", namespace, name)}},
	}, nil
}

func (a *KubernetesAdapter) getPodLogs(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	namespace := a.getNamespace(args)
	container, _ := args["container"].(string)
	tailLines := getIntArg(args, "tail_lines", 100)

	logs, err := a.client.GetPodLogs(ctx, namespace, name, container, tailLines)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: logs}},
	}, nil
}

func (a *KubernetesAdapter) execInPod(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	namespace := a.getNamespace(args)
	container, _ := args["container"].(string)
	cmdRaw, _ := args["command"].([]interface{})

	var cmd []string
	for _, c := range cmdRaw {
		if s, ok := c.(string); ok {
			cmd = append(cmd, s)
		}
	}

	output, err := a.client.ExecInPod(ctx, namespace, name, container, cmd)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: output}},
	}, nil
}

func (a *KubernetesAdapter) listDeployments(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	namespace := a.getNamespace(args)

	deployments, err := a.client.ListDeployments(ctx, namespace)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Deployments in '%s' (%d):\n\n", namespace, len(deployments)))

	for _, d := range deployments {
		sb.WriteString(fmt.Sprintf("- %s\n", d.Name))
		sb.WriteString(fmt.Sprintf("  Replicas: %d/%d ready, Image: %s\n", d.ReadyReplicas, d.Replicas, d.Image))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *KubernetesAdapter) scaleDeployment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	namespace := a.getNamespace(args)
	replicas := getIntArg(args, "replicas", 1)

	err := a.client.ScaleDeployment(ctx, namespace, name, replicas)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Scaled deployment %s/%s to %d replicas", namespace, name, replicas)}},
	}, nil
}

func (a *KubernetesAdapter) restartDeployment(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	name, _ := args["name"].(string)
	namespace := a.getNamespace(args)

	err := a.client.RestartDeployment(ctx, namespace, name)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Restarted deployment %s/%s", namespace, name)}},
	}, nil
}

func (a *KubernetesAdapter) listServices(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	namespace := a.getNamespace(args)

	services, err := a.client.ListServices(ctx, namespace)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Services in '%s' (%d):\n\n", namespace, len(services)))

	for _, s := range services {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", s.Name, s.Type))
		sb.WriteString(fmt.Sprintf("  ClusterIP: %s\n", s.ClusterIP))
		for _, p := range s.Ports {
			sb.WriteString(fmt.Sprintf("  Port: %d -> %d/%s\n", p.Port, p.TargetPort, p.Protocol))
		}
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *KubernetesAdapter) listNamespaces(ctx context.Context) (*ToolResult, error) {
	namespaces, err := a.client.ListNamespaces(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Namespaces (%d):\n\n", len(namespaces)))

	for _, ns := range namespaces {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", ns.Name, ns.Status))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *KubernetesAdapter) listNodes(ctx context.Context) (*ToolResult, error) {
	nodes, err := a.client.ListNodes(ctx)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Nodes (%d):\n\n", len(nodes)))

	for _, n := range nodes {
		sb.WriteString(fmt.Sprintf("- %s (%s)\n", n.Name, n.Status))
		sb.WriteString(fmt.Sprintf("  Roles: %s, Version: %s\n", strings.Join(n.Roles, ", "), n.Version))
		sb.WriteString(fmt.Sprintf("  CPU: %s, Memory: %s\n", n.CPUCapacity, n.MemoryCapacity))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *KubernetesAdapter) getEvents(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	namespace := a.getNamespace(args)
	fieldSelector, _ := args["field_selector"].(string)

	events, err := a.client.GetEvents(ctx, namespace, fieldSelector)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Events in '%s' (%d):\n\n", namespace, len(events)))

	for _, e := range events {
		icon := "ℹ️"
		if e.Type == "Warning" {
			icon = "⚠️"
		}
		sb.WriteString(fmt.Sprintf("%s %s: %s\n", icon, e.Reason, e.Message))
		sb.WriteString(fmt.Sprintf("   Object: %s, Count: %d\n", e.InvolvedObject, e.Count))
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: sb.String()}},
	}, nil
}

func (a *KubernetesAdapter) apply(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	yaml, _ := args["yaml"].(string)
	namespace := a.getNamespace(args)

	err := a.client.Apply(ctx, yaml, namespace)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: "Applied manifest successfully"}},
	}, nil
}

func (a *KubernetesAdapter) deleteResource(ctx context.Context, args map[string]interface{}) (*ToolResult, error) {
	kind, _ := args["kind"].(string)
	name, _ := args["name"].(string)
	namespace := a.getNamespace(args)

	err := a.client.Delete(ctx, kind, namespace, name)
	if err != nil {
		return &ToolResult{IsError: true, Content: []ContentBlock{{Type: "text", Text: err.Error()}}}, nil
	}

	return &ToolResult{
		Content: []ContentBlock{{Type: "text", Text: fmt.Sprintf("Deleted %s %s/%s", kind, namespace, name)}},
	}, nil
}
