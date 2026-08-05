package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"math/big"
	"os"
	"path/filepath"
	"testing"

	"github.com/bnb-chain/tss-lib/v2/crypto"
	"github.com/bnb-chain/tss-lib/v2/crypto/paillier"
	"github.com/bnb-chain/tss-lib/v2/ecdsa/keygen"
	"github.com/bnb-chain/tss-lib/v2/tss"

	"mpc/pkg/kms"
	"mpc/pkg/utils"
)

type fakeKMS struct {
	key []byte
}

func (f *fakeKMS) EncryptDEK(plaintextDEK []byte) ([]byte, error) {
	return utils.Encrypt(plaintextDEK, f.key)
}

func (f *fakeKMS) DecryptDEK(ciphertextDEK []byte) ([]byte, error) {
	return utils.Decrypt(ciphertextDEK, f.key)
}

func (f *fakeKMS) RotatedConfig() (kms.KeyManagementService, error) {
	return &fakeKMS{key: bytes.Repeat([]byte{0x55}, 32)}, nil
}

func newTestStore(t *testing.T, kmsSvc kms.KeyManagementService) *EncryptedLocalShareStore {
	t.Helper()
	store, err := NewEncryptedLocalShareStore(t.TempDir(), kmsSvc)
	if err != nil {
		t.Fatalf("NewEncryptedLocalShareStore: %v", err)
	}
	return store
}

func genPoint(t *testing.T) *crypto.ECPoint {
	t.Helper()
	gx, ok := new(big.Int).SetString("79BE667EF9DCBBAC55A06295CE870B07029BFCDB2DCE28D959F2815B16F81798", 16)
	if !ok {
		t.Fatal("bad Gx")
	}
	gy, ok := new(big.Int).SetString("483ADA7726A3C4655DA4FBFC0E1108A8FD17B448A68554199C47D08FFB10D4B8", 16)
	if !ok {
		t.Fatal("bad Gy")
	}
	pt, err := crypto.NewECPoint(tss.S256(), gx, gy)
	if err != nil {
		t.Fatalf("NewECPoint: %v", err)
	}
	return pt
}

func makeSaveData(t *testing.T) *keygen.LocalPartySaveData {
	t.Helper()
	sk, _, err := paillier.GenerateKeyPair(context.Background(), rand.Reader, 512)
	if err != nil {
		t.Fatalf("paillier keygen: %v", err)
	}

	pub := genPoint(t)

	return &keygen.LocalPartySaveData{
		LocalPreParams: keygen.LocalPreParams{
			PaillierSK: sk,
			NTildei:    big.NewInt(11),
			H1i:        big.NewInt(22),
			H2i:        big.NewInt(33),
		},
		LocalSecrets: keygen.LocalSecrets{
			Xi:      big.NewInt(123456789),
			ShareID: big.NewInt(987654321),
		},
		ECDSAPub: pub,
		BigXj:    []*crypto.ECPoint{pub},
	}
}

func TestSaveGetRoundTrip(t *testing.T) {
	store := newTestStore(t, &fakeKMS{key: []byte("01234567890123456789012345678901")})
	want := makeSaveData(t)

	if err := store.Save("wallet-1", want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	got, err := store.Get("wallet-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}

	if got.Xi == nil || got.Xi.Cmp(want.Xi) != 0 {
		t.Fatalf("Xi lost in round-trip: %v", got.Xi)
	}
	if got.ShareID == nil || got.ShareID.Cmp(want.ShareID) != 0 {
		t.Fatalf("ShareID lost in round-trip: %v", got.ShareID)
	}
	if got.PaillierSK == nil {
		t.Fatal("PaillierSK lost in round-trip")
	}
	if got.ECDSAPub == nil || got.ECDSAPub.X().Cmp(want.ECDSAPub.X()) != 0 {
		t.Fatal("ECDSAPub lost in round-trip")
	}
}

func TestEnvelopeLayoutOnDisk(t *testing.T) {
	masterKey := []byte("01234567890123456789012345678901")
	store := newTestStore(t, &fakeKMS{key: masterKey})

	if err := store.Save("wallet-1", makeSaveData(t)); err != nil {
		t.Fatalf("Save: %v", err)
	}

	blob, err := os.ReadFile(filepath.Join(store.baseDir, "wallet-1.share.enc"))
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if blob[0] != formatVersion {
		t.Fatalf("expected format version %d, got %d", formatVersion, blob[0])
	}

	if _, err := utils.Decrypt(blob[1:], masterKey); err == nil {
		t.Fatal("file should not be directly decryptable with the master key")
	}
}

func TestSwapRejected(t *testing.T) {
	store := newTestStore(t, &fakeKMS{key: []byte("01234567890123456789012345678901")})

	if err := store.Save("wallet-1", makeSaveData(t)); err != nil {
		t.Fatalf("Save: %v", err)
	}

	src := filepath.Join(store.baseDir, "wallet-1.share.enc")
	dst := filepath.Join(store.baseDir, "wallet-2.share.enc")
	if err := os.Rename(src, dst); err != nil {
		t.Fatalf("Rename: %v", err)
	}

	if _, err := store.Get("wallet-2"); err == nil {
		t.Fatal("Get on a swapped share file should fail (AAD mismatch)")
	}
}

func TestWrongKMSKeyRejected(t *testing.T) {
	keyA := []byte("01234567890123456789012345678901")
	keyB := []byte("abcdefghijklmnopqrstuvwxyz012345")

	storeA := newTestStore(t, &fakeKMS{key: keyA})
	if err := storeA.Save("wallet-1", makeSaveData(t)); err != nil {
		t.Fatalf("Save: %v", err)
	}

	storeB, err := NewEncryptedLocalShareStore(storeA.baseDir, &fakeKMS{key: keyB})
	if err != nil {
		t.Fatalf("NewEncryptedLocalShareStore: %v", err)
	}
	if _, err := storeB.Get("wallet-1"); err == nil {
		t.Fatal("Get with a different master key should fail")
	}
}

func TestRotateMasterKey(t *testing.T) {
	keyA := []byte("01234567890123456789012345678901")
	keyB := []byte("abcdefghijklmnopqrstuvwxyz012345")

	store := newTestStore(t, &fakeKMS{key: keyA})
	want := makeSaveData(t)
	if err := store.Save("wallet-1", want); err != nil {
		t.Fatalf("Save: %v", err)
	}

	newKMS := &fakeKMS{key: keyB}
	if err := store.RotateMasterKey(newKMS); err != nil {
		t.Fatalf("RotateMasterKey: %v", err)
	}

	got, err := store.Get("wallet-1")
	if err != nil {
		t.Fatalf("Get after rotation: %v", err)
	}
	if got.Xi == nil || got.Xi.Cmp(want.Xi) != 0 {
		t.Fatal("Xi lost after rotation")
	}

	oldStore, err := NewEncryptedLocalShareStore(store.baseDir, &fakeKMS{key: keyA})
	if err != nil {
		t.Fatalf("NewEncryptedLocalShareStore: %v", err)
	}
	if _, err := oldStore.Get("wallet-1"); err == nil {
		t.Fatal("old KMS should no longer read the file after rotation")
	}

	if err := store.RotateMasterKey(newKMS); err != nil {
		t.Fatalf("RotateMasterKey (repeat): %v", err)
	}
	if _, err := store.Get("wallet-1"); err != nil {
		t.Fatalf("Get after repeat rotation: %v", err)
	}
}

func TestRollbackMasterKey(t *testing.T) {
	keyA := []byte("01234567890123456789012345678901")
	keyB := []byte("abcdefghijklmnopqrstuvwxyz012345")

	storeA := newTestStore(t, &fakeKMS{key: keyA})
	want := makeSaveData(t)
	if err := storeA.Save("wallet-1", want); err != nil {
		t.Fatalf("Save: %v", err)
	}
	if err := storeA.RotateMasterKey(&fakeKMS{key: keyB}); err != nil {
		t.Fatalf("RotateMasterKey: %v", err)
	}

	recovered, err := NewEncryptedLocalShareStore(storeA.baseDir, &fakeKMS{key: keyA})
	if err != nil {
		t.Fatalf("NewEncryptedLocalShareStore: %v", err)
	}

	if err := recovered.RollbackMasterKey(&fakeKMS{key: keyB}, &fakeKMS{key: keyA}); err != nil {
		t.Fatalf("RollbackMasterKey: %v", err)
	}

	got, err := recovered.Get("wallet-1")
	if err != nil {
		t.Fatalf("Get after rollback: %v", err)
	}
	if got.Xi == nil || got.Xi.Cmp(want.Xi) != 0 {
		t.Fatal("Xi lost after rollback")
	}

	interimStore, err := NewEncryptedLocalShareStore(storeA.baseDir, &fakeKMS{key: keyB})
	if err != nil {
		t.Fatalf("NewEncryptedLocalShareStore: %v", err)
	}
	if _, err := interimStore.Get("wallet-1"); err == nil {
		t.Fatal("interim KMS should no longer read the file after rollback")
	}
}
