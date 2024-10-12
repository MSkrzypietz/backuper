package storage

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"
)

type Provider interface {
	GetName() string
	Backup(name string, zipPath string) error
}

type Storage struct {
	cfg       Config
	providers []Provider
}

func NewStorage(cfg Config) *Storage {
	return &Storage{
		cfg:       cfg,
		providers: make([]Provider, 0),
	}
}

func (s *Storage) GetAllProviders() []Provider {
	return s.providers
}

func (s *Storage) RegisterProvider(provider Provider) {
	if !s.HasProvider(provider.GetName()) {
		s.providers = append(s.providers, provider)
	}
}

func (s *Storage) HasProvider(name string) bool {
	_, ok := s.GetProvider(name)
	return ok
}

func (s *Storage) GetProvider(name string) (Provider, bool) {
	for _, provider := range s.providers {
		if provider.GetName() == name {
			return provider, true
		}
	}
	return nil, false
}

func (s *Storage) Backup(name string) error {
	cfg, ok := s.cfg.BackupConfigs[name]
	if !ok {
		return fmt.Errorf("unknown backup name: %s", name)
	}

	provider, ok := s.GetProvider(cfg.Provider)
	if !ok {
		return fmt.Errorf("unknown provider: %s", cfg.Provider)
	}

	backupName := fmt.Sprintf("backuper-%s-%s", name, time.Now().Format("2006-01-02_15-04-05"))
	zipDir := path.Join("/", "tmp", backupName)
	if err := zipDirectory(cfg.Path, zipDir); err != nil {
		return fmt.Errorf("error zipping the directory: %w", err)
	}
	return provider.Backup(backupName, zipDir)
}

func zipDirectory(sourceDir, targetZip string) error {
	zipFile, err := os.Create(targetZip)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	archive := zip.NewWriter(zipFile)
	defer archive.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		f, err := archive.Create(path)
		if err != nil {
			return err
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return err
		}

		return nil
	})

	return err
}
