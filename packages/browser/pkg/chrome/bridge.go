package chrome

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"

	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	BridgeName = "com.datakit.chrome_bridge"
	BridgeFile = BridgeName + ".json"
)

func GetBridgeFilePath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	switch runtime.GOOS {
	case "darwin":
		configDir = filepath.Join(configDir, "Google", "Chrome", "NativeMessagingHosts")
	case "linux":
		configDir = filepath.Join(configDir, "google-chrome", "NativeMessagingHosts")
	case "windows":
		// For Windows, user needs to manually add a registry entry pointing to this file.
		// This function just returns the recommended path for the manifest JSON.
		configDir = filepath.Join(configDir, "Google", "Chrome", "User Data", "NativeMessagingHosts")
	default:
		return "", fmt.Errorf("unsupported OS: %s", runtime.GOOS)
	}

	return filepath.Join(configDir, BridgeFile), nil
}

func InstallBridge(log *slog.Logger, extId string) error {
	execPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path for executable: %w", err)
	}

	execPath, err = filepath.Abs(execPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path for executable: %w", err)
	}

	filePath, err := GetBridgeFilePath()
	if err != nil {
		return err
	}

	if _, err := os.Stat(filePath); err != nil && os.IsNotExist(err) {
		if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
			return fmt.Errorf("failed to create bridge directory: %w", err)
		}
	}

	log.Info("Installing chrome native messaging bridge",
		slog.String("extId", extId),
		slog.String("filePath", filePath),
		slog.String("execPath", execPath),
	)

	data, err := protojson.MarshalOptions{
		UseProtoNames: true,
		Multiline:     true,
		Indent:        "  ",
	}.Marshal(&browserv1beta.ChromeBridge{
		Name:        BridgeName,
		Description: "DataKit Chrome Runtime",
		Path:        execPath,
		Type:        "stdio",
		AllowedOrigins: []string{
			fmt.Sprintf("chrome-extension://%s/", extId),
		},
	})
	if err != nil {
		return fmt.Errorf("failed to marshal chrome bridge: %w", err)
	}

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return fmt.Errorf("failed to write chrome bridge: %w", err)
	}

	return nil
}
