package pkg

import (
	"fmt"
	"os"
	"path/filepath"

	browserv1beta "github.com/datakit-dev/dtkt-integrations/browser/pkg/proto/integration/browser/v1beta"
	"github.com/datakit-dev/dtkt-sdk/sdk-go/encoding"
)

const ConfigSubDir = "dtkt-browser"

func GetConfigDir() (string, error) {
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}

	configDir := filepath.Join(userConfigDir, ConfigSubDir)
	err = os.MkdirAll(configDir, 0755)
	if err != nil {
		return "", err
	}

	return configDir, nil
}

func GetConfigPath(extId string) (string, error) {
	configDir, err := GetConfigDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(configDir, fmt.Sprintf("config-%s.yaml", extId)), nil
}

func ReadConfig(configDir, extId string) (*browserv1beta.Config, error) {
	path := filepath.Join(configDir, fmt.Sprintf("config-%s.yaml", extId))
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := new(browserv1beta.Config)
	err = encoding.FromYAMLV2(data, config)
	if err != nil {
		return nil, err
	}

	return config, nil
}

func WriteConfig(configDir string, config *browserv1beta.Config) error {
	var extId string
	if config.GetChrome() != nil {
		extId = config.GetChrome().GetExtensionId()
	} else {
		return fmt.Errorf("config missing extension id")
	}

	data, err := encoding.ToYAMLV2(config)
	if err != nil {
		return err
	}

	file, err := os.Create(filepath.Join(configDir, fmt.Sprintf("config-%s.yaml", extId)))
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.Write(data)
	return err
}
