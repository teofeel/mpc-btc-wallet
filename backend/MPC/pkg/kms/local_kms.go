package kms

import (
	"crypto/hkdf"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mpc/pkg/utils"
	"os"
	"path/filepath"
	"strings"
)

const (
	masterKeySize = 32
	kekInfo       = "wallet-share-kek"
)

type LocalKMSConfig struct {
	MasterKeyHex string `json:"masterKeyHex"`
	SecretHex    string `json:"secretHex"`
}

type LocalFileKMS struct {
	masterKey  []byte
	secret     []byte
	kek        []byte
	configPath string
}

func NewLocalFileKMS(jsonFilePath string) (*LocalFileKMS, error) {
	data, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return nil, err
	}

	var cfg LocalKMSConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	masterKey, err := hex.DecodeString(cfg.MasterKeyHex)
	if err != nil {
		return nil, fmt.Errorf("[MPC]{KMS} master key is not valid: %w", err)
	}
	if len(masterKey) != masterKeySize {
		return nil, fmt.Errorf("[MPC]{KMS} master key must be %d bytes", masterKeySize)
	}

	if cfg.SecretHex == "" {
		secret := make([]byte, masterKeySize)
		if _, err := io.ReadFull(rand.Reader, secret); err != nil {
			return nil, err
		}
		cfg.SecretHex = hex.EncodeToString(secret)
		if err := persistConfig(cfg, jsonFilePath); err != nil {
			return nil, err
		}
	}

	secret, err := hex.DecodeString(cfg.SecretHex)
	if err != nil {
		return nil, fmt.Errorf("[MPC]{KMS} secret is not valid: %w", err)
	}
	if len(secret) != masterKeySize {
		return nil, fmt.Errorf("[MPC]{KMS} secret must be %d bytes", masterKeySize)
	}

	kek, err := hkdf.Key(sha256.New, masterKey, secret, kekInfo, masterKeySize)
	if err != nil {
		return nil, fmt.Errorf("[MPC]{KMS} couldnt derive KEK: %w", err)
	}

	return &LocalFileKMS{
		masterKey:  masterKey,
		secret:     secret,
		kek:        kek,
		configPath: jsonFilePath,
	}, nil
}

func (k *LocalFileKMS) EncryptDEK(plaintextDEK []byte) ([]byte, error) {
	return utils.Encrypt(plaintextDEK, k.kek)
}

func (k *LocalFileKMS) DecryptDEK(ciphertextDEK []byte) ([]byte, error) {
	return utils.Decrypt(ciphertextDEK, k.kek)
}

func canonicalConfigPath(path string) string {
	return strings.TrimSuffix(path, ".rotated.json")
}

func RotatedConfigFilePath(canonicalPath string) string {
	return canonicalPath + ".rotated.json"
}

func (k *LocalFileKMS) RotatedConfig() (KeyManagementService, error) {
	secret := make([]byte, masterKeySize)
	if _, err := io.ReadFull(rand.Reader, secret); err != nil {
		return nil, err
	}

	canonical := canonicalConfigPath(k.configPath)
	rotatedPath := RotatedConfigFilePath(canonical)

	if err := persistConfig(LocalKMSConfig{
		MasterKeyHex: hex.EncodeToString(k.masterKey),
		SecretHex:    hex.EncodeToString(secret),
	}, rotatedPath); err != nil {
		return nil, err
	}

	return NewLocalFileKMS(rotatedPath)
}

func (k *LocalFileKMS) CommitRotation() error {
	canonical := canonicalConfigPath(k.configPath)
	if canonical == k.configPath {
		return nil
	}

	if _, err := os.Stat(k.configPath); errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("[MPC]{KMS} commit rotation: rotated config %s no longer exists", k.configPath)
	}

	if err := persistConfig(LocalKMSConfig{
		MasterKeyHex: hex.EncodeToString(k.masterKey),
		SecretHex:    hex.EncodeToString(k.secret),
	}, canonical); err != nil {
		return fmt.Errorf("[MPC]{KMS} commit rotation: persist problem: %w", err)
	}

	_ = os.Remove(k.configPath)

	k.configPath = canonical
	return nil
}

func (k *LocalFileKMS) ConfigFilePath() string {
	return canonicalConfigPath(k.configPath)
}

func persistConfig(cfg LocalKMSConfig, path string) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	tmp := filepath.Join(filepath.Dir(path), filepath.Base(path)+".tmp")
	if err := os.WriteFile(tmp, data, 0600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}
