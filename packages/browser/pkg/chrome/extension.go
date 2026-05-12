package chrome

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/pkg/browser"
)

const (
	LogFile         = "chrome.log"
	ExtensionSubDir = "chrome"
)

//go:embed extension/dist/*
var extFS embed.FS

//go:embed template/*
var tmplFS embed.FS

func GetExtensionIdFromArg(extArg string) string {
	return strings.TrimSuffix(strings.TrimPrefix(extArg, "chrome-extension://"), "/")
}

func InstallExtension(log *slog.Logger, configDir string) error {
	var (
		extDir       = filepath.Join(configDir, ExtensionSubDir)
		firstInstall bool
	)
	if stat, err := os.Stat(extDir); err != nil && os.IsNotExist(err) {
		err = os.MkdirAll(extDir, 0755)
		if err != nil {
			return fmt.Errorf("failed to create extension directory (%s): %s", extDir, err)
		}

		firstInstall = true
	} else if !stat.IsDir() {
		return fmt.Errorf("not a directory: %s", extDir)
	} else {
		// Clean directory before copying extension files.
		entries, err := os.ReadDir(extDir)
		if err != nil {
			return err
		}

		for _, entry := range entries {
			err := os.RemoveAll(filepath.Join(extDir, entry.Name()))
			if err != nil {
				return err
			}
		}
	}

	log.Info("Installing chrome extension", slog.String("extDir", extDir))

	err := fs.WalkDir(extFS, "extension", func(readPath string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		writePath := filepath.Join(extDir, strings.TrimPrefix(readPath, "extension/dist"))
		if d.IsDir() {
			return os.MkdirAll(writePath, 0755)
		}

		srcFile, err := extFS.Open(readPath)
		if err != nil {
			return err
		}

		dstFile, err := os.Create(writePath)
		if err != nil {
			return err
		}

		_, err = io.Copy(dstFile, srcFile)
		return err
	})
	if err != nil {
		return err
	}

	if firstInstall {
		tmpl, err := template.ParseFS(tmplFS, "template/install.html.tmpl")
		if err != nil {
			return err
		}

		var buf bytes.Buffer
		err = tmpl.Execute(&buf, map[string]any{
			"INSTALL_DIR": extDir,
		})
		if err != nil {
			return err
		}

		return browser.OpenReader(&buf)
	}

	return nil
}
