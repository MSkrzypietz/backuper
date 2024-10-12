package storage

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"io"
	"os"
	"path"
)

type BackupConfig struct {
	Path     string `yaml:"path"`
	Provider string `yaml:"provider"`
}

type ProviderConfig struct {
	AccessKey  string `yaml:"accessKey"`
	SecretKey  string `yaml:"secretKey"`
	AccountID  string `yaml:"accountID"`
	BucketName string `yaml:"bucketName"`
}

type Config struct {
	BackupConfigs   map[string]BackupConfig   `yaml:"backups"`
	ProviderConfigs map[string]ProviderConfig `yaml:"providers"`
}

func NewConfig(r io.Reader) (Config, error) {
	var config Config
	err := yaml.NewDecoder(r).Decode(&config)
	if err != nil {
		return config, err
	}
	return config, nil
}

func NewConfigFromEnv() (Config, error) {
	var config Config
	homeConfig := path.Join(os.Getenv("HOME"), ".backuper.yml")
	file, err := os.Open(homeConfig)
	if err != nil {
		return config, err
	}
	defer file.Close()
	return NewConfig(file)
}

func (c Config) Validate() error {
	for _, b := range c.BackupConfigs {
		if !c.HasProvider(b.Provider) {
			return fmt.Errorf("unknown provider: %s", b.Provider)
		}
		if !isValidPath(b.Path) {
			return fmt.Errorf("ErrInvalidPath: %s", b.Path)
		}
	}
	return nil
}

func isValidPath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func (c Config) HasProvider(name string) bool {
	for providerName := range c.ProviderConfigs {
		if providerName == name {
			return true
		}
	}
	return false
}
