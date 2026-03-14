package executor

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/kevinburke/ssh_config"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/agent"
	"golang.org/x/crypto/ssh/knownhosts"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	commandv1beta1 "github.com/datakit-dev/dtkt-sdk/sdk-go/proto/dtkt/command/v1beta1"
)

// applySSHConfigDefaults loads the SSH config file (if specified or using default)
// and applies defaults from it to the SSHConfig proto. Proto settings take precedence.
func applySSHConfigDefaults(sshConfig *commandv1beta1.SSHConfig) error {
	configPath := sshConfig.ConfigFilePath

	// Use default SSH config if not specified
	if configPath == "" {
		home, err := os.UserHomeDir()
		if err == nil {
			configPath = filepath.Join(home, ".ssh", "config")
		} else {
			return nil // Can't determine home dir, skip config loading
		}
	}

	// Expand path (handle ~ and env vars)
	configPath = os.ExpandEnv(configPath)
	if strings.HasPrefix(configPath, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to expand home directory: %w", err)
		}
		configPath = filepath.Join(home, configPath[2:])
	}

	// Load and parse config file
	f, err := os.Open(configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// Only allow missing file if using default path
			if sshConfig.ConfigFilePath != "" {
				return fmt.Errorf("specified SSH config file not found: %s", configPath)
			}
			return nil // Default config file doesn't exist, that's ok
		}
		return fmt.Errorf("failed to open SSH config file: %w", err)
	}
	//nolint:errcheck
	defer f.Close()

	cfg, err := ssh_config.Decode(f)
	if err != nil {
		return fmt.Errorf("failed to parse SSH config file: %w", err)
	}

	// Extract host from address for config lookup
	host, _, err := net.SplitHostPort(sshConfig.Address)
	if err != nil {
		// If no port, use the whole address as host
		host = sshConfig.Address
	}

	// Apply defaults from SSH config (proto settings override)

	// User
	if sshConfig.User == "" {
		if user, err := cfg.Get(host, "User"); err == nil && user != "" {
			sshConfig.User = user
		}
	}

	// ProxyCommand (only if not already set in proto)
	if sshConfig.ProxyCommand == "" {
		if proxyCmd, err := cfg.Get(host, "ProxyCommand"); err == nil && proxyCmd != "" && proxyCmd != "none" {
			sshConfig.ProxyCommand = proxyCmd
		}
	}

	// IdentityFile (only if no auth method is set)
	if sshConfig.Auth == nil {
		if identityFiles, err := cfg.GetAll(host, "IdentityFile"); err == nil && len(identityFiles) > 0 {
			// Use first existing identity file (SSH tries multiple, we simplify to one)
			identityPath := os.ExpandEnv(identityFiles[0])
			if strings.HasPrefix(identityPath, "~/") {
				home, err := os.UserHomeDir()
				if err == nil {
					identityPath = filepath.Join(home, identityPath[2:])
				}
			}

			// Check if file exists before setting
			if _, err := os.Stat(identityPath); err == nil {
				sshConfig.Auth = &commandv1beta1.SSHConfig_IdentityFile{
					IdentityFile: &commandv1beta1.SSHConfig_IdentityFileAuth{
						FilePath: identityPath,
					},
				}
			}
		}
	}

	// UserKnownHostsFile (only if using KnownHosts verification)
	if khVerify, ok := sshConfig.HostKeyVerification.(*commandv1beta1.SSHConfig_KnownHosts_); ok {
		if len(khVerify.KnownHosts.FilePaths) == 0 {
			if knownHostsFiles, err := cfg.GetAll(host, "UserKnownHostsFile"); err == nil && len(knownHostsFiles) > 0 {
				expanded := make([]string, 0, len(knownHostsFiles))
				for _, kh := range knownHostsFiles {
					khPath := os.ExpandEnv(kh)
					if strings.HasPrefix(khPath, "~/") {
						home, err := os.UserHomeDir()
						if err == nil {
							khPath = filepath.Join(home, khPath[2:])
						}
					}
					if _, err := os.Stat(khPath); err == nil {
						expanded = append(expanded, khPath)
					}
				}
				if len(expanded) > 0 {
					khVerify.KnownHosts.FilePaths = expanded
				}
			}
		}
	}

	return nil
}

// dialSSH establishes an SSH connection using the provided configuration.
// It handles both direct connections and proxy commands.
func dialSSH(ctx context.Context, sshConfig *commandv1beta1.SSHConfig, clientConfig *ssh.ClientConfig) (*ssh.Client, error) {
	var conn net.Conn
	var err error

	if sshConfig.ProxyCommand != "" {
		conn, err = executeProxyCommand(ctx, sshConfig, clientConfig)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to execute proxy command: %v", err)
		}
	} else {
		var dialer net.Dialer
		conn, err = dialer.DialContext(ctx, "tcp", sshConfig.Address)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to dial SSH server %s: %v", sshConfig.Address, err)
		}
	}

	// Establish SSH connection over the transport
	sshConn, chans, reqs, err := ssh.NewClientConn(conn, sshConfig.Address, clientConfig)
	if err != nil {
		//nolint:errcheck
		conn.Close()
		return nil, status.Errorf(codes.Internal, "failed to establish SSH connection to %s@%s: %v", clientConfig.User, sshConfig.Address, err)
	}

	return ssh.NewClient(sshConn, chans, reqs), nil
}

// buildSSHClientConfig creates an ssh.ClientConfig from SSHConfig proto.
func buildSSHClientConfig(sshConfig *commandv1beta1.SSHConfig) (*ssh.ClientConfig, error) {
	// Apply defaults from SSH config file (~/.ssh/config or custom path)
	if err := applySSHConfigDefaults(sshConfig); err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "failed to load SSH config: %v", err)
	}

	clientConfig := &ssh.ClientConfig{
		User: sshConfig.User,
	}

	// Handle authentication
	authMethods, err := buildAuthMethods(sshConfig)
	if err != nil {
		return nil, err
	}
	clientConfig.Auth = authMethods

	// Handle host key verification
	hostKeyCallback, err := buildHostKeyCallback(sshConfig)
	if err != nil {
		return nil, err
	}
	clientConfig.HostKeyCallback = hostKeyCallback

	return clientConfig, nil
}

// buildAuthMethods creates SSH authentication methods from the proto configuration.
// Supports private keys, identity files, passwords, and SSH agent (default).
func buildAuthMethods(sshConfig *commandv1beta1.SSHConfig) ([]ssh.AuthMethod, error) {
	switch auth := sshConfig.Auth.(type) {
	case *commandv1beta1.SSHConfig_PrivateKey:
		var signer ssh.Signer
		var err error

		if auth.PrivateKey.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(auth.PrivateKey.KeyData, []byte(auth.PrivateKey.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(auth.PrivateKey.KeyData)
		}

		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse private key: %v", err)
		}

		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil

	case *commandv1beta1.SSHConfig_IdentityFile:
		keyData, err := os.ReadFile(auth.IdentityFile.FilePath)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to read identity file: %v", err)
		}

		var signer ssh.Signer
		if auth.IdentityFile.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyData, []byte(auth.IdentityFile.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(keyData)
		}

		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to parse identity file: %v", err)
		}

		return []ssh.AuthMethod{ssh.PublicKeys(signer)}, nil

	case *commandv1beta1.SSHConfig_Password:
		return []ssh.AuthMethod{ssh.Password(auth.Password)}, nil

	case nil:
		// Default to SSH agent
		agentConn, err := net.Dial("unix", os.Getenv("SSH_AUTH_SOCK"))
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "SSH agent not available (SSH_AUTH_SOCK not set or connection failed): %v", err)
		}
		agentClient := agent.NewClient(agentConn)
		return []ssh.AuthMethod{ssh.PublicKeysCallback(agentClient.Signers)}, nil

	default:
		return nil, status.Errorf(codes.InvalidArgument, "unsupported authentication method: %T", auth)
	}
}

// buildHostKeyCallback creates the host key verification callback.
// Supports TOFU, known_hosts files, fingerprint verification, and insecure skip.
func buildHostKeyCallback(sshConfig *commandv1beta1.SSHConfig) (ssh.HostKeyCallback, error) {
	switch verification := sshConfig.HostKeyVerification.(type) {
	case *commandv1beta1.SSHConfig_TofuFilePath:
		// Trust On First Use: accept and store key on first connection, verify on subsequent
		tofuPath := verification.TofuFilePath
		if tofuPath == "" {
			// Use default TOFU path
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to get home directory: %v", err)
			}
			tofuPath = filepath.Join(home, ".ssh", "known_hosts_tofu")
		}

		return makeTOFUCallback(tofuPath), nil

	case *commandv1beta1.SSHConfig_KnownHosts_:
		knownHostsPaths := verification.KnownHosts.FilePaths
		if len(knownHostsPaths) == 0 {
			home, err := os.UserHomeDir()
			if err != nil {
				return nil, status.Errorf(codes.Internal, "failed to get home directory: %v", err)
			}
			knownHostsPaths = []string{
				filepath.Join(home, ".ssh", "known_hosts"),
			}
		}

		callback, err := makeKnownHostsCallback(knownHostsPaths)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "failed to create known_hosts callback: %v", err)
		}
		return callback, nil

	case *commandv1beta1.SSHConfig_Fingerprint:
		return makeFingerprintCallback(verification.Fingerprint), nil

	case *commandv1beta1.SSHConfig_InsecureSkipVerify:
		if verification.InsecureSkipVerify {
			return ssh.InsecureIgnoreHostKey(), nil
		}
		// If false, fall through to default

	case nil:
		// Default to TOFU with default path
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to get home directory: %v", err)
		}
		tofuPath := filepath.Join(home, ".ssh", "known_hosts_tofu")
		return makeTOFUCallback(tofuPath), nil
	}

	return nil, status.Error(codes.InvalidArgument, "invalid host key verification configuration")
}

func makeTOFUCallback(tofuPath string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		// Ensure directory exists
		if err := os.MkdirAll(filepath.Dir(tofuPath), 0700); err != nil {
			return fmt.Errorf("failed to create TOFU directory: %w", err)
		}

		// Normalize hostname to handle [host]:port format
		normalizedHost := knownhosts.Normalize(hostname)

		// Check if host key already exists
		storedKey, err := loadHostKey(tofuPath, normalizedHost)
		if err == nil {
			// Key exists, verify it matches
			if !bytes.Equal(storedKey.Marshal(), key.Marshal()) {
				return fmt.Errorf("host key mismatch for %s: expected %s, got %s",
					normalizedHost,
					ssh.FingerprintSHA256(storedKey),
					ssh.FingerprintSHA256(key))
			}
			return nil
		}

		// Key doesn't exist, store it (first use)
		if err := storeHostKey(tofuPath, normalizedHost, key); err != nil {
			return fmt.Errorf("failed to store host key: %w", err)
		}

		return nil
	}
}

func makeKnownHostsCallback(knownHostsPaths []string) (ssh.HostKeyCallback, error) {
	var existingPaths []string
	for _, path := range knownHostsPaths {
		if _, err := os.Stat(path); err == nil {
			existingPaths = append(existingPaths, path)
		}
	}

	if len(existingPaths) == 0 {
		return nil, fmt.Errorf("no known_hosts files found")
	}

	return knownhosts.New(existingPaths...)
}

func makeFingerprintCallback(expectedFingerprint string) ssh.HostKeyCallback {
	return func(hostname string, remote net.Addr, key ssh.PublicKey) error {
		actualFingerprint := ssh.FingerprintSHA256(key)

		// Try both SHA256 and legacy MD5 formats
		if actualFingerprint != expectedFingerprint {
			// Try MD5 format
			actualFingerprintMD5 := ssh.FingerprintLegacyMD5(key)
			if actualFingerprintMD5 != expectedFingerprint {
				return fmt.Errorf("host key fingerprint mismatch for %s: expected %s, got %s (SHA256) or %s (MD5)",
					hostname, expectedFingerprint, actualFingerprint, actualFingerprintMD5)
			}
		}

		return nil
	}
}

func loadHostKey(path, hostname string) (ssh.PublicKey, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Fields(line)
		if len(parts) < 3 {
			continue
		}

		if parts[0] == hostname {
			keyData, err := base64.StdEncoding.DecodeString(parts[2])
			if err != nil {
				continue
			}
			key, err := ssh.ParsePublicKey(keyData)
			if err != nil {
				continue
			}
			return key, nil
		}
	}

	return nil, fmt.Errorf("host key not found for %s", hostname)
}

func storeHostKey(path, hostname string, key ssh.PublicKey) error {
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}
	//nolint:errcheck
	defer f.Close()

	line := fmt.Sprintf("%s %s %s\n", hostname, key.Type(), base64.StdEncoding.EncodeToString(key.Marshal()))
	_, err = f.WriteString(line)
	return err
}

// executeProxyCommand executes the proxy command and returns a connection using its stdin/stdout.
func executeProxyCommand(ctx context.Context, sshConfig *commandv1beta1.SSHConfig, clientConfig *ssh.ClientConfig) (net.Conn, error) {
	host, port, err := net.SplitHostPort(sshConfig.Address)
	if err != nil {
		return nil, fmt.Errorf("failed to parse address: %w", err)
	}

	// Expand %h (host), %p (port), %r (user) variables
	cmdStr := sshConfig.ProxyCommand
	cmdStr = strings.ReplaceAll(cmdStr, "%h", host)
	cmdStr = strings.ReplaceAll(cmdStr, "%p", port)
	cmdStr = strings.ReplaceAll(cmdStr, "%r", clientConfig.User)

	// Execute command using shell with context for cancellation
	cmd := exec.CommandContext(ctx, "sh", "-c", cmdStr)

	// Capture stderr for error messages
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start proxy command: %w (stderr: %s)", err, stderrBuf.String())
	}

	// Create a connection from the pipes
	return &proxyConn{
		cmd:       cmd,
		stdin:     stdin,
		stdout:    stdout,
		stderrBuf: &stderrBuf,
	}, nil
}

// proxyConn implements net.Conn by tunneling through a proxy command's stdin/stdout.
type proxyConn struct {
	cmd       *exec.Cmd
	stdin     io.WriteCloser
	stdout    io.ReadCloser
	stderrBuf *bytes.Buffer
}

// Ensure proxyConn implements net.Conn
var _ net.Conn = (*proxyConn)(nil)

func (c *proxyConn) Read(b []byte) (n int, err error) {
	return c.stdout.Read(b)
}

func (c *proxyConn) Write(b []byte) (n int, err error) {
	return c.stdin.Write(b)
}

func (c *proxyConn) Close() error {
	// Close both pipes - this signals the proxy command to exit
	stdinErr := c.stdin.Close()
	stdoutErr := c.stdout.Close()

	// Wait for process to exit naturally
	waitErr := c.cmd.Wait()

	// Return first error encountered
	if stdinErr != nil {
		return fmt.Errorf("failed to close stdin: %w", stdinErr)
	}
	if stdoutErr != nil {
		return fmt.Errorf("failed to close stdout: %w", stdoutErr)
	}
	if waitErr != nil {
		return fmt.Errorf("proxy command failed: %w (stderr: %s)", waitErr, c.stderrBuf.String())
	}
	return nil
}

func (c *proxyConn) LocalAddr() net.Addr {
	return &net.TCPAddr{}
}

func (c *proxyConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{}
}

func (c *proxyConn) SetDeadline(t time.Time) error {
	// Deadlines are not supported for proxy command connections
	if !t.IsZero() {
		return fmt.Errorf("deadlines not supported for proxy command connections")
	}
	return nil
}

func (c *proxyConn) SetReadDeadline(t time.Time) error {
	if !t.IsZero() {
		return fmt.Errorf("deadlines not supported for proxy command connections")
	}
	return nil
}

func (c *proxyConn) SetWriteDeadline(t time.Time) error {
	if !t.IsZero() {
		return fmt.Errorf("deadlines not supported for proxy command connections")
	}
	return nil
}
