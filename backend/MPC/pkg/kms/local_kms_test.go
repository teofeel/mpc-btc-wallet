package kms

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func testMasterKey() []byte {
	return []byte("01234567890123456789012345678901") // 32 bytes
}

func testSecret() []byte {
	return []byte("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa") // 32 bytes
}

func writeConfig(t *testing.T, cfg LocalKMSConfig) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "kms.json")
	data, err := json.Marshal(cfg)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	return path
}

func loadKMS(t *testing.T, cfg LocalKMSConfig) *LocalFileKMS {
	t.Helper()
	k, err := NewLocalFileKMS(writeConfig(t, cfg))
	if err != nil {
		t.Fatalf("NewLocalFileKMS: %v", err)
	}
	return k
}

func testConfig() LocalKMSConfig {
	return LocalKMSConfig{
		MasterKeyHex: hex.EncodeToString(testMasterKey()),
		SecretHex:    hex.EncodeToString(testSecret()),
	}
}

func TestWrapRoundTrip(t *testing.T) {
	k := loadKMS(t, testConfig())

	dek := bytes.Repeat([]byte{0x42}, 32)
	wrapped, err := k.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}

	got, err := k.DecryptDEK(wrapped)
	if err != nil {
		t.Fatalf("DecryptDEK: %v", err)
	}

	if !bytes.Equal(got, dek) {
		t.Fatalf("DEK mismatch: got %x", got)
	}
}

func TestSameConfigUnwrap(t *testing.T) {
	cfg := testConfig()
	k1 := loadKMS(t, cfg)

	dek := bytes.Repeat([]byte{0x42}, 32)
	wrapped, err := k1.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}

	k2 := loadKMS(t, cfg)
	got, err := k2.DecryptDEK(wrapped)
	if err != nil {
		t.Fatalf("DecryptDEK with fresh instance: %v", err)
	}
	if !bytes.Equal(got, dek) {
		t.Fatalf("DEK mismatch: got %x", got)
	}
}

func TestWrongMasterKeyRejected(t *testing.T) {
	k1 := loadKMS(t, testConfig())

	dek := bytes.Repeat([]byte{0x42}, 32)
	wrapped, err := k1.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}

	other := testConfig()
	other.MasterKeyHex = hex.EncodeToString(bytes.Repeat([]byte{0x11}, 32))
	k2 := loadKMS(t, other)

	if _, err := k2.DecryptDEK(wrapped); err == nil {
		t.Fatal("DecryptDEK with a different master key should fail")
	}
}

func TestWrongSecretRejected(t *testing.T) {
	k1 := loadKMS(t, testConfig())

	dek := bytes.Repeat([]byte{0x42}, 32)
	wrapped, err := k1.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}

	other := testConfig()
	other.SecretHex = hex.EncodeToString(bytes.Repeat([]byte{0x22}, 32))
	k2 := loadKMS(t, other)

	if _, err := k2.DecryptDEK(wrapped); err == nil {
		t.Fatal("DecryptDEK with a different secret should fail")
	}
}

func TestRotatedConfig(t *testing.T) {
	k1 := loadKMS(t, testConfig())

	dek := bytes.Repeat([]byte{0x42}, 32)
	wrapped, err := k1.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}

	rotated, err := k1.RotatedConfig()
	if err != nil {
		t.Fatalf("RotatedConfig: %v", err)
	}

	k2, ok := rotated.(*LocalFileKMS)
	if !ok {
		t.Fatalf("RotatedConfig must return *LocalFileKMS, got %T", rotated)
	}
	if _, err := k2.DecryptDEK(wrapped); err == nil {
		t.Fatal("rotated KEK must not unwrap old blobs")
	}

	data, err := os.ReadFile(k2.configPath)
	if err != nil {
		t.Fatalf("ReadFile rotated config: %v", err)
	}
	var cfg LocalKMSConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		t.Fatalf("Unmarshal rotated config: %v", err)
	}
	if cfg.MasterKeyHex != testConfig().MasterKeyHex {
		t.Fatal("RotatedConfig must keep the master key")
	}
	if cfg.SecretHex == testConfig().SecretHex {
		t.Fatal("RotatedConfig must mint a fresh secret")
	}

	rotWrapped, err := k2.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}
	k3, err := NewLocalFileKMS(k2.configPath)
	if err != nil {
		t.Fatalf("NewLocalFileKMS after rotation: %v", err)
	}
	got, err := k3.DecryptDEK(rotWrapped)
	if err != nil {
		t.Fatalf("DecryptDEK after reload: %v", err)
	}
	if !bytes.Equal(got, dek) {
		t.Fatalf("DEK mismatch: got %x", got)
	}
}

func TestBootstrapMintsSecret(t *testing.T) {
	cfg := LocalKMSConfig{MasterKeyHex: hex.EncodeToString(testMasterKey())}
	path := writeConfig(t, cfg)

	k1, err := NewLocalFileKMS(path)
	if err != nil {
		t.Fatalf("NewLocalFileKMS: %v", err)
	}

	dek := bytes.Repeat([]byte{0x42}, 32)
	wrapped, err := k1.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}

	k2, err := NewLocalFileKMS(path)
	if err != nil {
		t.Fatalf("NewLocalFileKMS after bootstrap: %v", err)
	}
	got, err := k2.DecryptDEK(wrapped)
	if err != nil {
		t.Fatalf("DecryptDEK after restart: %v", err)
	}
	if !bytes.Equal(got, dek) {
		t.Fatalf("DEK mismatch: got %x", got)
	}
}

func TestMasterKeyLengthValidated(t *testing.T) {
	cfg := testConfig()
	cfg.MasterKeyHex = hex.EncodeToString([]byte("short"))
	if _, err := NewLocalFileKMS(writeConfig(t, cfg)); err == nil {
		t.Fatal("NewLocalFileKMS with a short master key should fail")
	}
}

func TestSecretLengthValidated(t *testing.T) {
	cfg := testConfig()
	cfg.SecretHex = hex.EncodeToString([]byte("short"))
	if _, err := NewLocalFileKMS(writeConfig(t, cfg)); err == nil {
		t.Fatal("NewLocalFileKMS with a short secret should fail")
	}
}

func TestCommitRotation(t *testing.T) {
	original := testConfig()
	cfgPath := writeConfig(t, original)

	k1, err := NewLocalFileKMS(cfgPath)
	if err != nil {
		t.Fatalf("NewLocalFileKMS: %v", err)
	}

	if k1.ConfigFilePath() != cfgPath {
		t.Fatalf("ConfigFilePath: expected %s, got %s", cfgPath, k1.ConfigFilePath())
	}

	rotated, err := k1.RotatedConfig()
	if err != nil {
		t.Fatalf("RotatedConfig: %v", err)
	}

	k2 := rotated.(*LocalFileKMS)
	rotatedPath := RotatedConfigFilePath(cfgPath)

	if _, err := os.Stat(rotatedPath); os.IsNotExist(err) {
		t.Fatal("rotated config file should exist before commit")
	}

	if err := k2.CommitRotation(); err != nil {
		t.Fatalf("CommitRotation: %v", err)
	}

	if _, err := os.Stat(rotatedPath); !os.IsNotExist(err) {
		t.Fatal("rotated config file should be removed after commit")
	}

	if k2.ConfigFilePath() != cfgPath {
		t.Fatalf("ConfigFilePath after commit: expected %s, got %s", cfgPath, k2.ConfigFilePath())
	}

	k3, err := NewLocalFileKMS(cfgPath)
	if err != nil {
		t.Fatalf("NewLocalFileKMS after commit: %v", err)
	}

	dek := bytes.Repeat([]byte{0x42}, 32)

	wrappedOld, err := k1.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK (old): %v", err)
	}

	if _, err := k3.DecryptDEK(wrappedOld); err == nil {
		t.Fatal("committed KMS must not decrypt blobs wrapped by the old KEK")
	}

	wrappedNew, err := k3.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK (new): %v", err)
	}

	got, err := k3.DecryptDEK(wrappedNew)
	if err != nil {
		t.Fatalf("DecryptDEK (new): %v", err)
	}

	if !bytes.Equal(got, dek) {
		t.Fatal("DEK mismatch after commit round-trip")
	}
}

func TestCommitRotationIdempotent(t *testing.T) {
	cfgPath := writeConfig(t, testConfig())
	k, err := NewLocalFileKMS(cfgPath)
	if err != nil {
		t.Fatalf("NewLocalFileKMS: %v", err)
	}

	if err := k.CommitRotation(); err != nil {
		t.Fatalf("CommitRotation on canonical KMS should be a no-op: %v", err)
	}

	if k.ConfigFilePath() != cfgPath {
		t.Fatal("ConfigFilePath changed after no-op commit")
	}

	dek := bytes.Repeat([]byte{0x42}, 32)
	wrapped, err := k.EncryptDEK(dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}
	got, err := k.DecryptDEK(wrapped)
	if err != nil {
		t.Fatalf("DecryptDEK: %v", err)
	}
	if !bytes.Equal(got, dek) {
		t.Fatal("DEK mismatch after no-op commit")
	}
}
