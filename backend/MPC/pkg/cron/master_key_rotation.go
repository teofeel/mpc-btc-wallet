package cron

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/robfig/cron/v3"

	"mpc/pkg/kms"
	"mpc/pkg/storage"
)

type KeyRotationManager struct {
	kms   *kms.LocalFileKMS
	store *storage.EncryptedLocalShareStore
	cron  *cron.Cron
}

func NewKeyRotationManager(localKMS *kms.LocalFileKMS, store *storage.EncryptedLocalShareStore) *KeyRotationManager {
	return &KeyRotationManager{
		kms:   localKMS,
		store: store,
		cron:  cron.New(cron.WithSeconds()),
	}
}

func (m *KeyRotationManager) RecoverOnStartup() error {
	rotatedPath := kms.RotatedConfigFilePath(m.kms.ConfigFilePath())
	if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
		return nil
	}

	log.Println("[MPC]{KMS Cron} Found pending rotated config, attempting recovery...")

	rotatedKMS, err := kms.NewLocalFileKMS(rotatedPath)
	if err != nil {
		log.Printf("[MPC]{KMS Cron} Corrupt rotated config, discarding: %v", err)

		return os.Remove(rotatedPath)
	}

	migrated, err := m.sharesReadableWith(rotatedKMS)
	if err != nil {
		log.Printf("[MPC]{KMS Cron} Recovery check failed: %v - discarding", err)

		if rmErr := os.Remove(rotatedPath); rmErr != nil {
			return fmt.Errorf("[MPC]{KMS Cron} recovery: check data: %w; cleanup rotated: %w", err, rmErr)
		}

		return nil
	}

	if migrated {
		log.Println("[MPC]{KMS Cron} Shares already migrated - completing rotation")

		if err := rotatedKMS.CommitRotation(); err != nil {
			return fmt.Errorf("[MPC]{KMS Cron} recovery: commit rotation: %w", err)
		}

		m.kms = rotatedKMS
		log.Println("[MPC]{KMS Cron} Rotation completed successfully")
	} else {
		log.Println("[MPC]{KMS Cron} Data not yet migrated - discarding rotated config")

		if err := os.Remove(rotatedPath); err != nil {
			return fmt.Errorf("[MPC]{KMS Cron} recovery: cleanup rotated config: %w", err)
		}
	}

	return nil
}

func (m *KeyRotationManager) sharesReadableWith(testKMS *kms.LocalFileKMS) (bool, error) {
	tmpStore, err := storage.NewEncryptedLocalShareStore(m.store.BaseDir(), testKMS)
	if err != nil {
		return false, err
	}

	entries, err := os.ReadDir(m.store.BaseDir())
	if err != nil {
		return false, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".share.enc") {
			continue
		}
		walletID := strings.TrimSuffix(entry.Name(), ".share.enc")
		_, err := tmpStore.Get(walletID)
		if err == nil {
			return true, nil
		}

		return false, nil
	}

	return true, nil
}

func (m *KeyRotationManager) StartQuarterlyRotation() error {
	_, err := m.cron.AddFunc("0 0 0 1 */3 *", func() {
		log.Println("[MPC]{KMS Cron} Starting master key rotation...")

		if err := m.executeRotation(); err != nil {
			log.Printf("[MPC]{KMS Cron} Rotation failed: %v", err)
			return
		}

		log.Println("[MPC]{KMS Cron} Master key rotation completed")
	})

	if err != nil {
		return fmt.Errorf("[MPC]{KMS Cron} cron: schedule master key rotation: %w", err)
	}

	m.cron.Start()
	log.Println("[MPC]{KMS Cron} Master key rotation scheduled")
	return nil
}

func (m *KeyRotationManager) executeRotation() error {
	oldKMS := m.kms

	rotatedKMSSvc, err := m.kms.RotatedConfig()
	if err != nil {
		return fmt.Errorf("[MPC]{KMS Cron} rotation: generate new config: %w", err)
	}

	rotatedKMS, ok := rotatedKMSSvc.(*kms.LocalFileKMS)
	if !ok {
		return fmt.Errorf("[MPC]{KMS Cron} rotation: expected *LocalFileKMS from RotatedConfig")
	}

	if err := m.store.RotateMasterKey(rotatedKMS); err != nil {
		if rmErr := os.Remove(kms.RotatedConfigFilePath(oldKMS.ConfigFilePath())); rmErr != nil {
			log.Printf("[MPC]{KMS Cron} failed to clean up rotated config after migration error: %v", rmErr)
		}
		return fmt.Errorf("[MPC]{KMS Cron} rotation: migrate DEKs: %w", err)
	}

	if err := rotatedKMS.CommitRotation(); err != nil {
		return fmt.Errorf("[MPC]{KMS Cron} rotation: commit config: %w", err)
	}

	m.kms = rotatedKMS
	log.Println("[MPC]{KMS Cron} Master key rotation committed successfully")
	return nil
}

func (m *KeyRotationManager) Stop() {
	ctx := m.cron.Stop()
	<-ctx.Done()
}
