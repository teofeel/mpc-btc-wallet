package kms

type KeyManagementService interface {
	EncryptDEK(plaintextDEK []byte) ([]byte, error)
	DecryptDEK(ciphertextDEK []byte) ([]byte, error)
	RotatedConfig() (KeyManagementService, error)
}
