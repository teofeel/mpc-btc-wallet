package storage

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"mpc/pkg/kms"
	"mpc/pkg/utils"

	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
)

const formatVersion byte = 0x01

var _ PartyShareStore = (*EncryptedLocalShareStore)(nil)

type EncryptedLocalShareStore struct {
	mu      sync.RWMutex
	baseDir string
	kms     kms.KeyManagementService
}

func NewEncryptedLocalShareStore(baseDir string, kms kms.KeyManagementService) (*EncryptedLocalShareStore, error) {
	if kms == nil {
		return nil, errors.New("[MPC]{LocalShareStorage} storage: kms must not be null")
	}
	if err := os.MkdirAll(baseDir, 0700); err != nil {
		return nil, err
	}

	return &EncryptedLocalShareStore{baseDir: baseDir, kms: kms}, nil
}

func (s *EncryptedLocalShareStore) getFilePath(walletID string) string {
	return filepath.Join(s.baseDir, fmt.Sprintf("%s.share.enc", walletID))
}

func (s *EncryptedLocalShareStore) BaseDir() string {
	return s.baseDir
}

func (s *EncryptedLocalShareStore) encrypt(walletID string, plaintext []byte) ([]byte, error) {
	dek := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, dek); err != nil {
		return nil, err
	}

	wrappedDEK, err := s.kms.EncryptDEK(dek)
	if err != nil {
		return nil, err
	}

	payload, err := utils.Encrypt(plaintext, dek, []byte(walletID))
	if err != nil {
		return nil, err
	}

	blob := make([]byte, 0, 1+4+len(wrappedDEK)+len(payload))
	blob = append(blob, formatVersion)
	blob = binary.BigEndian.AppendUint32(blob, uint32(len(wrappedDEK)))
	blob = append(blob, wrappedDEK...)

	return append(blob, payload...), nil
}

func (s *EncryptedLocalShareStore) decrypt(walletID string, blob []byte) ([]byte, error) {
	if len(blob) < 1+4 {
		return nil, errors.New("[MPC]{LocalShareStorage} storage: share file too short")
	}
	if blob[0] != formatVersion {
		return nil, fmt.Errorf("[MPC]{LocalShareStorage} storage: unsupported share file version %d", blob[0])
	}

	wrappedLen := binary.BigEndian.Uint32(blob[1:5])
	if int(wrappedLen) > len(blob)-5 {
		return nil, errors.New("[MPC]{LocalShareStorage} storage: corrupt share file (wrapped DEK length)")
	}

	wrappedDEK := blob[5 : 5+int(wrappedLen)]
	payload := blob[5+int(wrappedLen):]

	dek, err := s.kms.DecryptDEK(wrappedDEK)
	if err != nil {
		return nil, fmt.Errorf("[MPC]{LocalShareStorage} storage: unwrap DEK: %w", err)
	}

	return utils.Decrypt(payload, dek, []byte(walletID))
}

func (s *EncryptedLocalShareStore) Save(walletID string, saveData *keygen.LocalPartySaveData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	plaintext, err := json.Marshal(saveData)
	if err != nil {
		return err
	}

	blob, err := s.encrypt(walletID, plaintext)
	if err != nil {
		return err
	}

	tmpFile := filepath.Join(s.baseDir, fmt.Sprintf("%s.share.enc.tmp", walletID))
	if err := os.WriteFile(tmpFile, blob, 0600); err != nil {
		return err
	}

	return os.Rename(tmpFile, s.getFilePath(walletID))
}

func (s *EncryptedLocalShareStore) SaveAll(walletID string, saveDataByShare map[string]*keygen.LocalPartySaveData) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for shareID, saveData := range saveDataByShare {
		plaintext, err := json.Marshal(saveData)
		if err != nil {
			return err
		}

		blob, err := s.encrypt(walletID+":"+shareID, plaintext)
		if err != nil {
			return err
		}

		tmpFile := filepath.Join(s.baseDir, fmt.Sprintf("%s:%s.share.enc.tmp", walletID, shareID))
		if err := os.WriteFile(tmpFile, blob, 0600); err != nil {
			return err
		}

		if err := os.Rename(tmpFile, s.getFilePath(walletID+":"+shareID)); err != nil {
			return err
		}
	}

	return nil
}

func (s *EncryptedLocalShareStore) Get(walletID string) (*keygen.LocalPartySaveData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.readShare(walletID)
}

func (s *EncryptedLocalShareStore) GetAll(walletID string) (map[string]*keygen.LocalPartySaveData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]*keygen.LocalPartySaveData)

	if own, err := s.readShare(walletID); err == nil {
		result[own.ShareID.String()] = own
	} else if !os.IsNotExist(err) {
		return nil, err
	}

	prefix := walletID + ":"
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), prefix) || !strings.HasSuffix(entry.Name(), ".share.enc") {
			continue
		}

		key := strings.TrimSuffix(strings.TrimPrefix(entry.Name(), prefix), ".share.enc")
		saveData, err := s.readShare(walletID + ":" + key)
		if err != nil {
			return nil, err
		}

		result[saveData.ShareID.String()] = saveData
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("no key shares found for wallet %s", walletID)
	}

	return result, nil
}

func (s *EncryptedLocalShareStore) readShare(walletID string) (*keygen.LocalPartySaveData, error) {
	ciphertext, err := os.ReadFile(s.getFilePath(walletID))
	if err != nil {
		return nil, err
	}

	plaintext, err := s.decrypt(walletID, ciphertext)
	if err != nil {
		return nil, err
	}

	var saveData keygen.LocalPartySaveData
	if err := json.Unmarshal(plaintext, &saveData); err != nil {
		return nil, err
	}

	return &saveData, nil
}

func (s *EncryptedLocalShareStore) RotateMasterKey(newKMS kms.KeyManagementService) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.migrate(s.kms, newKMS); err != nil {
		return err
	}

	s.kms = newKMS
	return nil
}

func (s *EncryptedLocalShareStore) RollbackMasterKey(interimKMS, lastGoodKMS kms.KeyManagementService) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.migrate(interimKMS, lastGoodKMS); err != nil {
		return err
	}

	s.kms = lastGoodKMS
	return nil
}

func (s *EncryptedLocalShareStore) migrate(fromKMS, toKMS kms.KeyManagementService) error {
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return err
	}

	type stagedFile struct {
		path string
		blob []byte
	}

	var staged []stagedFile
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".share.enc") {
			continue
		}

		path := filepath.Join(s.baseDir, entry.Name())
		blob, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if len(blob) < 5 || blob[0] != formatVersion {
			return fmt.Errorf("[MPC]{LocalShareStorage} storage: unsupported data file %s", entry.Name())
		}

		wrappedLen := binary.BigEndian.Uint32(blob[1:5])
		if int(wrappedLen) > len(blob)-5 {
			return fmt.Errorf("[MPC]{LocalShareStorage} storage: corrupt data file %s", entry.Name())
		}

		wrappedDEK := blob[5 : 5+int(wrappedLen)]
		payload := blob[5+int(wrappedLen):]

		dek, err := fromKMS.DecryptDEK(wrappedDEK)
		if err != nil {
			if _, err2 := toKMS.DecryptDEK(wrappedDEK); err2 == nil {
				continue
			}
			return fmt.Errorf("[MPC]{LocalShareStorage} storage: cannot unwrap %s: %w", entry.Name(), err)
		}

		newWrapped, err := toKMS.EncryptDEK(dek)
		if err != nil {
			return err
		}

		newBlob := make([]byte, 0, 5+len(newWrapped)+len(payload))
		newBlob = append(newBlob, formatVersion)
		newBlob = binary.BigEndian.AppendUint32(newBlob, uint32(len(newWrapped)))
		newBlob = append(newBlob, newWrapped...)
		newBlob = append(newBlob, payload...)

		staged = append(staged, stagedFile{path: path, blob: newBlob})
	}

	for _, f := range staged {
		tmp := f.path + ".tmp"
		if err := os.WriteFile(tmp, f.blob, 0600); err != nil {
			return err
		}
		if err := os.Rename(tmp, f.path); err != nil {
			return err
		}
	}

	return nil
}

func (s *EncryptedLocalShareStore) Delete(walletID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	return os.Remove(s.getFilePath(walletID))
}

func (s *EncryptedLocalShareStore) Exists(walletID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, err := os.Stat(s.getFilePath(walletID))
	return err == nil
}
