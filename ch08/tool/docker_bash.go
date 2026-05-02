package tool

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/openai/openai-go/v3"
	"github.com/openai/openai-go/v3/shared"
)

const (
	DefaultSandboxContainer = "babyagent-sandbox"
	DefaultSandboxImage     = "alpine:3.19"
)

// generateContainerName generates a unique container name based on workspace directory
func generateContainerName(workspaceDir string) string {
	// Use the project directory name as suffix
	projectName := filepath.Base(workspaceDir)
	if projectName == "" || projectName == "." || projectName == "/" {
		return DefaultSandboxContainer
	}
	return fmt.Sprintf("%s-%s", DefaultSandboxContainer, projectName)
}

type DockerBashTool struct {
	containerName string
	image         string
	workspaceDir  string

	once      sync.Once
	startErr  error
	eventHook EventHook
}

func NewDockerBashTool(containerName, workspaceDir string) *DockerBashTool {
	if containerName == "" {
		containerName = generateContainerName(workspaceDir)
	}
	return &DockerBashTool{
		containerName: containerName,
		image:         DefaultSandboxImage,
		workspaceDir:  workspaceDir,
	}
}

func (t *DockerBashTool) ToolName() AgentTool {
	return AgentToolBash
}

func (t *DockerBashTool) RuntimeDescription() string {
	return fmt.Sprintf("bash tool: docker sandbox container=%s image=%s workspace=%s mount=/workspace:rw", t.containerName, t.image, t.workspaceDir)
}

func (t *DockerBashTool) SetEventHook(hook EventHook) {
	t.eventHook = hook
}

func (t *DockerBashTool) emit(stage, message string) {
	if t.eventHook == nil {
		return
	}
	t.eventHook(ToolEvent{
		ToolName: string(t.ToolName()),
		Stage:    stage,
		Message:  message,
	})
}

func (t *DockerBashTool) Info() openai.ChatCompletionToolUnionParam {
	return openai.ChatCompletionFunctionTool(shared.FunctionDefinitionParam{
		Name:        string(AgentToolBash),
		Description: openai.String("execute bash command in a docker sandbox container"),
		Parameters: openai.FunctionParameters{
			"type": "object",
			"properties": map[string]any{
				"command": map[string]any{
					"type":        "string",
					"description": "the bash command to execute in the sandbox",
				},
			},
			"required": []string{"command"},
		},
	})
}

func (t *DockerBashTool) Execute(ctx context.Context, argumentsInJSON string) (string, error) {
	// Lazy initialization: start container on first use
	t.once.Do(func() {
		t.emit("sandbox_init_start", fmt.Sprintf("ensuring docker sandbox container=%s image=%s", t.containerName, t.image))
		t.startErr = t.ensureSandboxContainer(ctx)
		if t.startErr != nil {
			t.emit("sandbox_init_failed", t.startErr.Error())
			return
		}
		t.emit("sandbox_init_done", fmt.Sprintf("docker sandbox ready container=%s", t.containerName))
	})
	if t.startErr != nil {
		return "", fmt.Errorf("failed to start sandbox container: %w", t.startErr)
	}

	p := BashToolParam{}
	err := json.Unmarshal([]byte(argumentsInJSON), &p)
	if err != nil {
		return "", err
	}

	// Execute command in container via docker exec
	cmd := exec.CommandContext(ctx, "docker", "exec",
		t.containerName,
		"sh", "-c", p.Command)

	start := time.Now()
	t.emit("execute_start", fmt.Sprintf("docker exec %s sh -c %q", t.containerName, p.Command))
	output, err := cmd.CombinedOutput()
	t.emit("execute_done", fmt.Sprintf("docker exec finished in %s, output_bytes=%d, err=%v", time.Since(start).Round(time.Millisecond), len(output), err))
	if err != nil {
		return string(output), fmt.Errorf("docker exec failed: %w", err)
	}
	return string(output), nil
}

func (t *DockerBashTool) ensureSandboxContainer(ctx context.Context) error {
	// First, try to start existing container
	t.emit("docker_start", fmt.Sprintf("docker start %s", t.containerName))
	startCmd := exec.CommandContext(ctx, "docker", "start", t.containerName)
	if startCmd.Run() == nil {
		// Container exists and started successfully
		t.emit("docker_start_done", fmt.Sprintf("reused existing container=%s", t.containerName))
		return nil
	}

	// Container doesn't exist, create new one
	t.emit("docker_create", fmt.Sprintf("docker run image=%s container=%s mount=%s:/workspace:rw", t.image, t.containerName, t.workspaceDir))
	createCmd := exec.CommandContext(ctx, "docker", "run", "-d",
		"--name", t.containerName,
		"--restart", "unless-stopped",
		"-v", t.workspaceDir+":/workspace:rw",
		"-w", "/workspace",
		t.image,
		"sleep", "infinity")

	output, err := createCmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create sandbox container: %s: %w", string(output), err)
	}
	t.emit("docker_create_done", fmt.Sprintf("created container=%s", t.containerName))
	return nil
}
