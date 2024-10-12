package main

import (
	"fmt"
	"github.com/MSkrzypietz/backuper/internal/cfr2"
	"github.com/MSkrzypietz/backuper/internal/logger"
	"github.com/MSkrzypietz/backuper/internal/storage"
)

func main() {
	config, err := storage.NewConfigFromEnv()
	if err != nil {
		logger.Fatal("failed to create config", "err", err)
	}

	err = config.Validate()
	if err != nil {
		logger.Fatal("invalid config", "err", err)
	}

	s := storage.NewStorage(config)
	provider, err := cfr2.NewCloudflareR2Provider(config.ProviderConfigs[cfr2.ProviderName])
	if err != nil {
		logger.Error("failed to create r2 provider", "err", err)
	} else {
		s.RegisterProvider(provider)
	}

	for backupName := range config.BackupConfigs {
		logger.Info(fmt.Sprintf("backing up %s...", backupName))
		err = s.Backup(backupName)
		if err != nil {
			logger.Error("failed to backup", "err", err)
		} else {
			logger.Info(fmt.Sprintf("Successfully uploaded all files for %s", backupName))
		}
	}
}
