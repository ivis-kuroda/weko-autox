package dockerx

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"

	composetypes "github.com/compose-spec/compose-go/types"
	"github.com/docker/docker/api/types"
	containertypes "github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type ExecResult struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

type Executor interface {
	Exec(ctx context.Context, shellCommand string) (ExecResult, error)
	ExecStream(ctx context.Context, shellCommand string, stdoutWriter io.Writer, stderrWriter io.Writer) (ExecResult, error)
	StopTests(ctx context.Context) error
}

type Client struct {
	docker      *client.Client
	containerID string
}

func New(ctx context.Context, workspace string, serviceName string) (*Client, error) {
	composeFile, err := DetectComposeFile(workspace)
	if err != nil {
		return nil, err
	}
	project, err := LoadProject(composeFile)
	if err != nil {
		return nil, err
	}
	if !hasService(project, serviceName) {
		return nil, fmt.Errorf("service %q not found in compose", serviceName)
	}

	dockerCli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("create docker client: %w", err)
	}

	containerID, err := findServiceContainer(ctx, dockerCli, project.Name, serviceName)
	if err != nil {
		_ = dockerCli.Close()
		return nil, err
	}

	return &Client{
		docker:      dockerCli,
		containerID: containerID,
	}, nil
}

func (c *Client) Close() error {
	if c.docker == nil {
		return nil
	}
	return c.docker.Close()
}

func (c *Client) Exec(ctx context.Context, shellCommand string) (ExecResult, error) {
	return c.ExecStream(ctx, shellCommand, nil, nil)
}

func (c *Client) ExecStream(ctx context.Context, shellCommand string, stdoutWriter io.Writer, stderrWriter io.Writer) (ExecResult, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer

	execResp, err := c.docker.ContainerExecCreate(ctx, c.containerID, containertypes.ExecOptions{
		Cmd:          []string{"sh", "-lc", shellCommand},
		AttachStdout: true,
		AttachStderr: true,
		Tty:          false,
	})
	if err != nil {
		return ExecResult{}, fmt.Errorf("create exec: %w", err)
	}

	attachResp, err := c.docker.ContainerExecAttach(ctx, execResp.ID, containertypes.ExecStartOptions{})
	if err != nil {
		return ExecResult{}, fmt.Errorf("attach exec: %w", err)
	}
	defer attachResp.Close()

	// Close the connection when ctx is cancelled so StdCopy unblocks immediately.
	stopWatch := make(chan struct{})
	defer close(stopWatch)
	go func() {
		select {
		case <-ctx.Done():
			attachResp.Close()
		case <-stopWatch:
		}
	}()

	stdoutDest := io.Writer(&stdout)
	if stdoutWriter != nil {
		stdoutDest = io.MultiWriter(&stdout, stdoutWriter)
	}
	stderrDest := io.Writer(&stderr)
	if stderrWriter != nil {
		stderrDest = io.MultiWriter(&stderr, stderrWriter)
	}

	if _, err := stdcopy.StdCopy(stdoutDest, stderrDest, attachResp.Reader); err != nil && err != io.EOF {
		if ctx.Err() != nil {
			return ExecResult{ExitCode: -1}, ctx.Err()
		}
		return ExecResult{}, fmt.Errorf("read exec output: %w", err)
	}

	if ctx.Err() != nil {
		return ExecResult{ExitCode: -1}, ctx.Err()
	}

	inspect, err := c.docker.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return ExecResult{}, fmt.Errorf("inspect exec: %w", err)
	}

	result := ExecResult{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: inspect.ExitCode,
	}
	if result.ExitCode != 0 {
		return result, fmt.Errorf("exec failed with code %d", result.ExitCode)
	}
	return result, nil
}

func (c *Client) StopTests(ctx context.Context) error {
	_, err := c.Exec(ctx, "pkill -f 'tox|pytest' || true")
	return err
}

func hasService(project *composetypes.Project, serviceName string) bool {
	for _, svc := range project.Services {
		if svc.Name == serviceName {
			return true
		}
	}
	return false
}

func findServiceContainer(ctx context.Context, dockerCli *client.Client, projectName string, serviceName string) (string, error) {
	labelFilters := filters.NewArgs(
		filters.Arg("label", "com.docker.compose.service="+serviceName),
	)
	if strings.TrimSpace(projectName) != "" {
		labelFilters.Add("label", "com.docker.compose.project="+projectName)
	}

	containers, err := dockerCli.ContainerList(ctx, containertypes.ListOptions{All: true, Filters: labelFilters})
	if err != nil {
		return "", fmt.Errorf("list containers: %w", err)
	}

	containerID, err := chooseContainerID(containers)
	if err != nil {
		return "", fmt.Errorf("no container found for compose service %q", serviceName)
	}
	return containerID, nil
}

func chooseContainerID(containers []types.Container) (string, error) {
	if len(containers) == 0 {
		return "", fmt.Errorf("empty container list")
	}

	for _, ctr := range containers {
		if strings.EqualFold(ctr.State, "running") {
			return ctr.ID, nil
		}
	}

	return containers[0].ID, nil
}
