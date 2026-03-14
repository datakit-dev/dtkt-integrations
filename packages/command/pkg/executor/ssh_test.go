package executor

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"golang.org/x/crypto/ssh"

	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
)

// sshContainer holds the SSH test container
type sshContainer struct {
	testcontainers.Container
	Host     string
	Port     string
	User     string
	Password string
}

// Package-level container for all SSH tests
var sshTestContainer *sshContainer

// TestMain handles setup and teardown for all tests in this package
func TestMain(m *testing.M) {
	// Run tests
	code := m.Run()

	// Cleanup: terminate the SSH container if it was started
	if sshTestContainer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		_ = sshTestContainer.Terminate(ctx)
	}

	os.Exit(code)
}

// checkDockerAvailable verifies that Docker daemon is running and accessible
func checkDockerAvailable(ctx context.Context) error {
	provider, err := testcontainers.NewDockerProvider()
	if err != nil {
		return fmt.Errorf("failed to create Docker provider: %w", err)
	}
	//nolint:errcheck
	defer provider.Close()

	// Try to ping the Docker daemon
	if _, err := provider.Client().Ping(ctx); err != nil {
		return fmt.Errorf("cannot connect to Docker daemon: %w", err)
	}
	return nil
}

// setupSSHContainer creates the SSH container if not already running
func setupSSHContainer(t *testing.T) *sshContainer {
	t.Helper()

	// If container already exists, return it
	if sshTestContainer != nil {
		return sshTestContainer
	}

	ctx := context.Background()

	// Check if Docker is available
	if err := checkDockerAvailable(ctx); err != nil {
		t.Fatalf("Docker is not available: %v\n\nPlease ensure Docker Desktop is running before running integration tests.\nYou can skip integration tests with: go test -short ./...", err)
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
		t.Fatalf("failed to start SSH container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		//nolint:errcheck
		container.Terminate(ctx)
		t.Fatalf("failed to get container host: %v", err)
	}

	mappedPort, err := container.MappedPort(ctx, "2222")
	if err != nil {
		//nolint:errcheck
		container.Terminate(ctx)
		t.Fatalf("failed to get container port: %v", err)
	}

	sshTestContainer = &sshContainer{
		Container: container,
		Host:      host,
		Port:      mappedPort.Port(),
		User:      user,
		Password:  password,
	}

	return sshTestContainer
}

func (c *sshContainer) Address() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func (c *sshContainer) SSHConfig() *commandv1beta1.SSHConfig {
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

// SSH dial tests

func TestDialSSH_Password(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)
	ctx := context.Background()

	sshConfig := container.SSHConfig()

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	client, err := dialSSH(ctx, sshConfig, clientConfig)
	if err != nil {
		t.Fatalf("dialSSH failed: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	// Verify connection works by creating a session
	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	//nolint:errcheck
	defer session.Close()

	output, err := session.CombinedOutput("echo hello")
	if err != nil {
		t.Fatalf("echo command failed: %v", err)
	}

	if string(output) != "hello\n" {
		t.Errorf("expected 'hello\\n', got '%s'", output)
	}
}

func TestDialSSH_InvalidAddress(t *testing.T) {
	ctx := context.Background()

	sshConfig := &commandv1beta1.SSHConfig{
		Address: "localhost:99999",
		User:    "test",
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	_, err = dialSSH(ctx, sshConfig, clientConfig)
	if err == nil {
		t.Error("expected error for invalid address")
	}
}

func TestDialSSH_WrongPassword(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)
	ctx := context.Background()

	sshConfig := &commandv1beta1.SSHConfig{
		Address: container.Address(),
		User:    container.User,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "wrongpassword",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	_, err = dialSSH(ctx, sshConfig, clientConfig)
	if err == nil {
		t.Error("expected error for wrong password")
	}
}

func TestDialSSH_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	sshConfig := &commandv1beta1.SSHConfig{
		Address: "localhost:22",
		User:    "test",
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	_, err = dialSSH(ctx, sshConfig, clientConfig)
	if err == nil {
		t.Error("expected error for cancelled context")
	}
}

// buildSSHClientConfig tests

func TestBuildSSHClientConfig_Password(t *testing.T) {
	sshConfig := &commandv1beta1.SSHConfig{
		Address: "localhost:22",
		User:    "testuser",
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "testpass",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	if clientConfig.User != "testuser" {
		t.Errorf("expected user 'testuser', got '%s'", clientConfig.User)
	}

	if len(clientConfig.Auth) != 1 {
		t.Errorf("expected 1 auth method, got %d", len(clientConfig.Auth))
	}
}

func TestBuildSSHClientConfig_PrivateKey(t *testing.T) {
	// Generate a fresh RSA key for testing
	privateKey, _ := generateTestKeyPair(t)

	sshConfig := &commandv1beta1.SSHConfig{
		Address: "localhost:22",
		User:    "testuser",
		Auth: &commandv1beta1.SSHConfig_PrivateKey{
			PrivateKey: &commandv1beta1.SSHConfig_PrivateKeyAuth{
				KeyData: []byte(privateKey),
			},
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	if clientConfig.User != "testuser" {
		t.Errorf("expected user 'testuser', got '%s'", clientConfig.User)
	}

	if len(clientConfig.Auth) != 1 {
		t.Errorf("expected 1 auth method, got %d", len(clientConfig.Auth))
	}
}

func TestBuildSSHClientConfig_InvalidPrivateKey(t *testing.T) {
	sshConfig := &commandv1beta1.SSHConfig{
		Address: "localhost:22",
		User:    "testuser",
		Auth: &commandv1beta1.SSHConfig_PrivateKey{
			PrivateKey: &commandv1beta1.SSHConfig_PrivateKeyAuth{
				KeyData: []byte("not a valid key"),
			},
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err == nil {
		t.Error("expected error for invalid private key")
	}
}

func TestBuildSSHClientConfig_IdentityFile_NotFound(t *testing.T) {
	sshConfig := &commandv1beta1.SSHConfig{
		Address: "localhost:22",
		User:    "testuser",
		Auth: &commandv1beta1.SSHConfig_IdentityFile{
			IdentityFile: &commandv1beta1.SSHConfig_IdentityFileAuth{
				FilePath: "/nonexistent/path/to/key",
			},
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err == nil {
		t.Error("expected error for missing identity file")
	}
}

// Host key verification tests

func TestBuildHostKeyCallback_InsecureSkip(t *testing.T) {
	sshConfig := &commandv1beta1.SSHConfig{
		Address: "localhost:22",
		User:    "test",
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	if clientConfig.HostKeyCallback == nil {
		t.Error("expected HostKeyCallback to be set")
	}
}

func TestBuildHostKeyCallback_Fingerprint(t *testing.T) {
	sshConfig := &commandv1beta1.SSHConfig{
		Address: "localhost:22",
		User:    "test",
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_Fingerprint{
			Fingerprint: "SHA256:xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	if clientConfig.HostKeyCallback == nil {
		t.Error("expected HostKeyCallback to be set")
	}
}

func TestBuildHostKeyCallback_KnownHosts_NotFound(t *testing.T) {
	sshConfig := &commandv1beta1.SSHConfig{
		Address: "localhost:22",
		User:    "test",
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_KnownHosts_{
			KnownHosts: &commandv1beta1.SSHConfig_KnownHosts{
				FilePaths: []string{"/nonexistent/known_hosts"},
			},
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err == nil {
		t.Error("expected error for missing known_hosts file")
	}
}

// SSH config file tests

func TestApplySSHConfigDefaults_User(t *testing.T) {
	// Create temp SSH config file
	configContent := `
Host testhost
    User configuser
    Port 2222
`
	configFile := createTempFile(t, "ssh_config", configContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address:        "testhost:22",
		ConfigFilePath: configFile,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	// User should be populated from config
	if sshConfig.User != "configuser" {
		t.Errorf("expected user 'configuser', got '%s'", sshConfig.User)
	}
}

func TestApplySSHConfigDefaults_UserNotOverridden(t *testing.T) {
	// Create temp SSH config file
	configContent := `
Host testhost
    User configuser
`
	configFile := createTempFile(t, "ssh_config", configContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address:        "testhost:22",
		User:           "protouser", // Set in proto, should not be overridden
		ConfigFilePath: configFile,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	// User should NOT be overridden
	if sshConfig.User != "protouser" {
		t.Errorf("expected user 'protouser' (proto takes precedence), got '%s'", sshConfig.User)
	}
}

func TestApplySSHConfigDefaults_ProxyCommand(t *testing.T) {
	configContent := `
Host jumphost
    ProxyCommand ssh -W %h:%p bastion.example.com
`
	configFile := createTempFile(t, "ssh_config", configContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address:        "jumphost:22",
		User:           "testuser",
		ConfigFilePath: configFile,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	if sshConfig.ProxyCommand != "ssh -W %h:%p bastion.example.com" {
		t.Errorf("expected ProxyCommand from config, got '%s'", sshConfig.ProxyCommand)
	}
}

func TestApplySSHConfigDefaults_IdentityFile(t *testing.T) {
	// Generate a fresh key pair for testing
	testKey, _ := generateTestKeyPair(t)
	identityFile := createTempFile(t, "id_rsa", testKey)

	configContent := fmt.Sprintf(`
Host keyhost
    IdentityFile %s
    User keyuser
`, identityFile)
	configFile := createTempFile(t, "ssh_config", configContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address:        "keyhost:22",
		ConfigFilePath: configFile,
		// No Auth set - should pick up IdentityFile from config
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	// Should have picked up identity file
	if auth, ok := sshConfig.Auth.(*commandv1beta1.SSHConfig_IdentityFile); ok {
		if auth.IdentityFile.FilePath != identityFile {
			t.Errorf("expected identity file '%s', got '%s'", identityFile, auth.IdentityFile.FilePath)
		}
	} else {
		t.Errorf("expected IdentityFile auth, got %T", sshConfig.Auth)
	}
}

func TestApplySSHConfigDefaults_WildcardHost(t *testing.T) {
	configContent := `
Host *
    User defaultuser
    ServerAliveInterval 60

Host *.example.com
    User exampleuser
`
	configFile := createTempFile(t, "ssh_config", configContent)

	// Test wildcard match
	sshConfig := &commandv1beta1.SSHConfig{
		Address:        "anyhost:22",
		ConfigFilePath: configFile,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	if sshConfig.User != "defaultuser" {
		t.Errorf("expected user 'defaultuser' from wildcard, got '%s'", sshConfig.User)
	}
}

func TestApplySSHConfigDefaults_ConfigFileNotFound(t *testing.T) {
	sshConfig := &commandv1beta1.SSHConfig{
		Address:        "testhost:22",
		User:           "testuser",
		ConfigFilePath: "/nonexistent/ssh_config",
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err == nil {
		t.Error("expected error for missing config file when path is explicitly specified")
	}
}

// Note: ssh_config library is lenient and does not error on invalid syntax,
// so we skip testing invalid config syntax validation.

func TestApplySSHConfigDefaults_MultipleHosts(t *testing.T) {
	configContent := `
Host host1
    User user1
    Port 22

Host host2
    User user2
    Port 2222

Host host3
    User user3
`
	configFile := createTempFile(t, "ssh_config", configContent)

	// Test host2 specifically
	sshConfig := &commandv1beta1.SSHConfig{
		Address:        "host2:2222",
		ConfigFilePath: configFile,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	if sshConfig.User != "user2" {
		t.Errorf("expected user 'user2', got '%s'", sshConfig.User)
	}
}

func TestApplySSHConfigDefaults_ProxyCommandNone(t *testing.T) {
	// SSH config files process in order - specific hosts should come before wildcards
	// to allow the specific host's settings to take precedence
	configContent := `
Host direct
    ProxyCommand none

Host *
    ProxyCommand ssh -W %h:%p bastion
`
	configFile := createTempFile(t, "ssh_config", configContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address:        "direct:22",
		User:           "testuser",
		ConfigFilePath: configFile,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	_, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	// ProxyCommand "none" should result in empty ProxyCommand
	if sshConfig.ProxyCommand != "" {
		t.Errorf("expected empty ProxyCommand for 'none', got '%s'", sshConfig.ProxyCommand)
	}
}

// Integration tests - actually connect using SSH config

func TestSSHConfig_ConnectWithConfigFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	// Create SSH config file pointing to the container
	// Use the actual address since we can't rely on hostname resolution
	configContent := fmt.Sprintf(`
Host %s
    User %s
`, container.Host, container.User)

	configFile := createTempFile(t, "ssh_config", configContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address:        container.Address(),
		ConfigFilePath: configFile,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: container.Password,
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	// Actually connect
	client, err := dialSSH(context.Background(), sshConfig, clientConfig)
	if err != nil {
		t.Fatalf("dialSSH failed: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	// Run a command to verify connection works
	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	//nolint:errcheck
	defer session.Close()

	output, err := session.CombinedOutput("echo connected")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if string(output) != "connected\n" {
		t.Errorf("expected 'connected\\n', got '%s'", output)
	}
}

func TestSSHConfig_ConnectWithUserFromConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	// Create SSH config with User - should be picked up (use actual host)
	configContent := fmt.Sprintf(`
Host %s
    User %s
`, container.Host, container.User)

	configFile := createTempFile(t, "ssh_config", configContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address:        container.Address(),
		ConfigFilePath: configFile,
		// No User set - should pick up from config
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: container.Password,
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	// Verify user was picked up from config
	if clientConfig.User != container.User {
		t.Fatalf("expected user '%s' from config, got '%s'", container.User, clientConfig.User)
	}

	// Actually connect to prove it works
	client, err := dialSSH(context.Background(), sshConfig, clientConfig)
	if err != nil {
		t.Fatalf("dialSSH failed: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	//nolint:errcheck
	defer session.Close()

	output, err := session.CombinedOutput("whoami")
	if err != nil {
		t.Fatalf("whoami failed: %v", err)
	}

	if string(output) != container.User+"\n" {
		t.Errorf("expected '%s\\n', got '%s'", container.User, output)
	}
}

func TestSSHConfig_ConnectWithWildcard(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	// Create SSH config with wildcard
	configContent := fmt.Sprintf(`
Host *
    User %s
`, container.User)

	configFile := createTempFile(t, "ssh_config", configContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address:        container.Address(),
		ConfigFilePath: configFile,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: container.Password,
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	client, err := dialSSH(context.Background(), sshConfig, clientConfig)
	if err != nil {
		t.Fatalf("dialSSH failed: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	//nolint:errcheck
	defer session.Close()

	output, err := session.CombinedOutput("echo wildcard-works")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if string(output) != "wildcard-works\n" {
		t.Errorf("expected 'wildcard-works\\n', got '%s'", output)
	}
}

// Integration test: Identity file authentication
func TestSSHConfig_ConnectWithIdentityFile(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	// Generate a key pair for testing
	privateKey, publicKey := generateTestKeyPair(t)

	// Add the public key to the container's authorized_keys
	addAuthorizedKey(t, container, publicKey)

	// Write the private key to a temp file
	keyFile := createTempFile(t, "id_test", privateKey)

	sshConfig := &commandv1beta1.SSHConfig{
		Address: container.Address(),
		User:    container.User,
		Auth: &commandv1beta1.SSHConfig_IdentityFile{
			IdentityFile: &commandv1beta1.SSHConfig_IdentityFileAuth{
				FilePath: keyFile,
			},
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	client, err := dialSSH(context.Background(), sshConfig, clientConfig)
	if err != nil {
		t.Fatalf("dialSSH failed: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	//nolint:errcheck
	defer session.Close()

	output, err := session.CombinedOutput("echo key-auth-works")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if string(output) != "key-auth-works\n" {
		t.Errorf("expected 'key-auth-works\\n', got '%s'", output)
	}
}

// Integration test: Private key (inline) authentication
func TestSSHConfig_ConnectWithPrivateKey(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	// Generate a key pair for testing
	privateKey, publicKey := generateTestKeyPair(t)

	// Add the public key to the container's authorized_keys
	addAuthorizedKey(t, container, publicKey)

	sshConfig := &commandv1beta1.SSHConfig{
		Address: container.Address(),
		User:    container.User,
		Auth: &commandv1beta1.SSHConfig_PrivateKey{
			PrivateKey: &commandv1beta1.SSHConfig_PrivateKeyAuth{
				KeyData: []byte(privateKey),
			},
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	client, err := dialSSH(context.Background(), sshConfig, clientConfig)
	if err != nil {
		t.Fatalf("dialSSH failed: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	//nolint:errcheck
	defer session.Close()

	output, err := session.CombinedOutput("echo inline-key-works")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if string(output) != "inline-key-works\n" {
		t.Errorf("expected 'inline-key-works\\n', got '%s'", output)
	}
}

// Integration test: Fingerprint verification
func TestSSHConfig_ConnectWithFingerprint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	// Get the actual fingerprint from the container
	fingerprint := getHostFingerprint(t, container)

	sshConfig := &commandv1beta1.SSHConfig{
		Address: container.Address(),
		User:    container.User,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: container.Password,
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_Fingerprint{
			Fingerprint: fingerprint,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	client, err := dialSSH(context.Background(), sshConfig, clientConfig)
	if err != nil {
		t.Fatalf("dialSSH failed: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	//nolint:errcheck
	defer session.Close()

	output, err := session.CombinedOutput("echo fingerprint-verified")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if string(output) != "fingerprint-verified\n" {
		t.Errorf("expected 'fingerprint-verified\\n', got '%s'", output)
	}
}

// Integration test: Fingerprint mismatch should fail
func TestSSHConfig_ConnectWithWrongFingerprint(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	sshConfig := &commandv1beta1.SSHConfig{
		Address: container.Address(),
		User:    container.User,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: container.Password,
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_Fingerprint{
			Fingerprint: "SHA256:invalidfingerprintxxxxxxxxxxxxxxxxxxxxxxx",
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	_, err = dialSSH(context.Background(), sshConfig, clientConfig)
	if err == nil {
		t.Error("expected error for wrong fingerprint, got nil")
	}
}

// Integration test: Known hosts file verification
func TestSSHConfig_ConnectWithKnownHosts(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	// Get the host key and create a known_hosts file
	knownHostsContent := getKnownHostsEntry(t, container)
	knownHostsFile := createTempFile(t, "known_hosts", knownHostsContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address: container.Address(),
		User:    container.User,
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: container.Password,
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_KnownHosts_{
			KnownHosts: &commandv1beta1.SSHConfig_KnownHosts{
				FilePaths: []string{knownHostsFile},
			},
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	client, err := dialSSH(context.Background(), sshConfig, clientConfig)
	if err != nil {
		t.Fatalf("dialSSH failed: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	//nolint:errcheck
	defer session.Close()

	output, err := session.CombinedOutput("echo known-hosts-verified")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if string(output) != "known-hosts-verified\n" {
		t.Errorf("expected 'known-hosts-verified\\n', got '%s'", output)
	}
}

// Integration test: Connection timeout
func TestDialSSH_Timeout(t *testing.T) {
	// Use a non-routable IP to simulate timeout
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	sshConfig := &commandv1beta1.SSHConfig{
		Address: "10.255.255.1:22", // Non-routable IP
		User:    "test",
		Auth: &commandv1beta1.SSHConfig_Password{
			Password: "test",
		},
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	_, err = dialSSH(ctx, sshConfig, clientConfig)
	if err == nil {
		t.Error("expected timeout error, got nil")
	}
	if ctx.Err() != context.DeadlineExceeded {
		t.Errorf("expected context deadline exceeded, got: %v", ctx.Err())
	}
}

// Integration test: Identity file from SSH config file
func TestSSHConfig_ConnectWithIdentityFileFromConfig(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	container := setupSSHContainer(t)

	// Generate a key pair for testing
	privateKey, publicKey := generateTestKeyPair(t)

	// Add the public key to the container's authorized_keys
	addAuthorizedKey(t, container, publicKey)

	// Write the private key to a temp file
	keyFile := createTempFile(t, "id_test", privateKey)

	// Create SSH config that references the identity file (use actual host)
	configContent := fmt.Sprintf(`
Host %s
    User %s
    IdentityFile %s
`, container.Host, container.User, keyFile)

	configFile := createTempFile(t, "ssh_config", configContent)

	sshConfig := &commandv1beta1.SSHConfig{
		Address:        container.Address(),
		ConfigFilePath: configFile,
		// No explicit auth - should use IdentityFile from config
		HostKeyVerification: &commandv1beta1.SSHConfig_InsecureSkipVerify{
			InsecureSkipVerify: true,
		},
	}

	clientConfig, err := buildSSHClientConfig(sshConfig)
	if err != nil {
		t.Fatalf("buildSSHClientConfig failed: %v", err)
	}

	client, err := dialSSH(context.Background(), sshConfig, clientConfig)
	if err != nil {
		t.Fatalf("dialSSH failed: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		t.Fatalf("NewSession failed: %v", err)
	}
	//nolint:errcheck
	defer session.Close()

	output, err := session.CombinedOutput("echo identity-from-config")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if string(output) != "identity-from-config\n" {
		t.Errorf("expected 'identity-from-config\\n', got '%s'", output)
	}
}

// Helper to create temp files for testing
func createTempFile(t *testing.T, name, content string) string {
	t.Helper()
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, name)
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	return path
}

// Helper to generate an RSA key pair for testing
func generateTestKeyPair(t *testing.T) (privateKey, publicKey string) {
	t.Helper()

	// Generate RSA key
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate RSA key: %v", err)
	}

	// Encode private key to PEM
	privateKeyPEM := pem.EncodeToMemory(&pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	})

	// Generate public key in OpenSSH format
	pub, err := ssh.NewPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("failed to create SSH public key: %v", err)
	}
	publicKeyBytes := ssh.MarshalAuthorizedKey(pub)

	return string(privateKeyPEM), string(publicKeyBytes)
}

// Helper to add a public key to the container's authorized_keys
func addAuthorizedKey(t *testing.T, container *sshContainer, publicKey string) {
	t.Helper()

	ctx := context.Background()

	// The linuxserver/openssh-server container stores keys in /config/.ssh/authorized_keys
	// We need to append our key there
	cmd := []string{"sh", "-c", fmt.Sprintf(
		"mkdir -p /config/.ssh && echo '%s' >> /config/.ssh/authorized_keys && chmod 600 /config/.ssh/authorized_keys",
		strings.TrimSpace(publicKey),
	)}

	exitCode, _, err := container.Exec(ctx, cmd)
	if err != nil {
		t.Fatalf("failed to exec in container: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("failed to add authorized key, exit code: %d", exitCode)
	}
}

// Helper to get the host fingerprint from the container
func getHostFingerprint(t *testing.T, container *sshContainer) string {
	t.Helper()

	// Connect to get the host key
	config := &ssh.ClientConfig{
		User: container.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(container.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	client, err := ssh.Dial("tcp", container.Address(), config)
	if err != nil {
		t.Fatalf("failed to dial for fingerprint: %v", err)
	}
	//nolint:errcheck
	defer client.Close()

	// Get the host key fingerprint
	hostKey := client.RemoteAddr()
	_ = hostKey // We need to get the actual host key

	// Actually, we need to capture the host key during connection
	var capturedKey ssh.PublicKey
	captureConfig := &ssh.ClientConfig{
		User: container.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(container.Password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			capturedKey = key
			return nil
		},
		Timeout: 5 * time.Second,
	}

	client2, err := ssh.Dial("tcp", container.Address(), captureConfig)
	if err != nil {
		t.Fatalf("failed to dial for key capture: %v", err)
	}
	//nolint:errcheck
	client2.Close()

	if capturedKey == nil {
		t.Fatal("failed to capture host key")
	}

	return ssh.FingerprintSHA256(capturedKey)
}

// Helper to get a known_hosts entry for the container
func getKnownHostsEntry(t *testing.T, container *sshContainer) string {
	t.Helper()

	var capturedKey ssh.PublicKey
	config := &ssh.ClientConfig{
		User: container.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(container.Password),
		},
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			capturedKey = key
			return nil
		},
		Timeout: 5 * time.Second,
	}

	client, err := ssh.Dial("tcp", container.Address(), config)
	if err != nil {
		t.Fatalf("failed to dial for known_hosts: %v", err)
	}
	//nolint:errcheck
	client.Close()

	if capturedKey == nil {
		t.Fatal("failed to capture host key")
	}

	// Format: [host]:port key-type base64-key
	return fmt.Sprintf("[%s]:%s %s",
		container.Host,
		container.Port,
		strings.TrimSpace(string(ssh.MarshalAuthorizedKey(capturedKey))),
	)
}
