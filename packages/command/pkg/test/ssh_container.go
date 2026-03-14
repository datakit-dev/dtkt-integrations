package test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
)

type SSHContainer struct {
	testcontainers.Container
	Host     string
	Port     string
	User     string
	Password string
}

var (
	sshContainerOnce     sync.Once
	sshContainerInstance *SSHContainer
	sshContainerErr      error
)

func (c *SSHContainer) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c *SSHContainer) SSHConfig() *commandv1beta1.SSHConfig {
	return &commandv1beta1.SSHConfig{
		Address: c.Address(),
		User:    c.User,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: c.Password,
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}
}

func (c *SSHContainer) Terminate(ctx context.Context) error {
	if c.Container != nil {
		return c.Container.Terminate(ctx)
	}
	return nil
}

// checkDockerAvailable verifies that Docker daemon is running and accessible.
func checkDockerAvailable(ctx context.Context) error {
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return fmt.Errorf("failed to create Docker provider: %w", err)
	}
	//nolint:errcheck
	defer provider.Close()

	if _, err := provider.Client().Ping(ctx); err != nil {
		return fmt.Errorf("cannot connect to Docker daemon: %w", err)
	}
	return nil
}

// SetupSSHContainer creates a shared SSH container for testing.
// The container is created once and reused across all tests.
// Call TerminateSSHContainer in TestMain to clean up.
func SetupSSHContainer(t *testing.T) *SSHContainer {
	t.Helper()

	sshContainerOnce.Do(func() {
		ctx := context.Background()

		if err := checkDockerAvailable(ctx); err != nil {
			sshContainerErr = fmt.Errorf("please ensure docker is running before running integration tests.\nYou can skip integration tests with: `go test -short ./...`\n\n\ndocker is not available: %w", err)
			return
		}

		user := "testuser"
		password := "testpass"

		req := testcontainers.ContainerRequest{
			Image:        "linuxserver/openssh-server:latest",
			ExposedPorts: []string{"2222/tcp"},
			Env: map[string]string{
				"PUID":            "1000",
				"PGID":            "1000",
				"TZ":              "UTC",
				"USER_NAME":       user,
				"USER_PASSWORD":   password,
				"PASSWORD_ACCESS": "true",
			},
			WaitingFor: wait.ForListeningPort("2222/tcp").WithStartupTimeout(60 * time.Second),
		}

		container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
			ContainerRequest: req,
			Started:          true,
		})
		if err != nil {
			sshContainerErr = fmt.Errorf("failed to start SSH container: %w", err)
			return
		}

		host, err := container.Host(ctx)
		if err != nil {
			//nolint:errcheck
			container.Terminate(ctx)
			sshContainerErr = fmt.Errorf("failed to get container host: %w", err)
			return
		}

		mappedPort, err := container.MappedPort(ctx, "2222")
		if err != nil {
			//nolint:errcheck
			container.Terminate(ctx)
			sshContainerErr = fmt.Errorf("failed to get container port: %w", err)
			return
		}

		sshContainerInstance = &SSHContainer{
			Container: container,
			Host:      host,
			Port:      mappedPort.Port(),
			User:      user,
			Password:  password,
		}
	})

	if sshContainerErr != nil {
		t.Fatal(sshContainerErr)
	}

	return sshContainerInstance
}

// TerminateSSHContainer terminates the shared SSH container.
// Call this from TestMain after tests complete.
func TerminateSSHContainer() {
	if sshContainerInstance != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = sshContainerInstance.Terminate(ctx)
		sshContainerInstance = nil
	}
}

func GetSSHContainer() *SSHContainer {
	return sshContainerInstance
}
