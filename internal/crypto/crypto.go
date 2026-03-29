package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"io"
	"os"
)

const (
	Version  byte = 0x01
	SaltLen       = 16
	NonceLen      = 12
	KeyLen        = 32
)

var (
	ErrInvalidVersion = errors.New("invalid file version")
	ErrDecryption     = errors.New("decryption failed")
	ErrInvalidKey     = errors.New("invalid key")
)

type encryptedFile struct {
	Version byte
	Salt    []byte
	Nonce   []byte
	Cipher  []byte
}

func GenerateKey() (string, error) {
	key := make([]byte, KeyLen)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("failed to generate key: %w", err)
	}
	return hex.EncodeToString(key), nil
}

func DeriveKey(passphrase string, salt []byte) []byte {
	hash := argon2.IDKey(
		[]byte(passphrase),
		salt,
		1,
		64*1024,
		4,
		KeyLen,
	)
	return hash
}

func DeriveKeyFromHex(hexKey string) ([]byte, error) {
	key, err := hex.DecodeString(hexKey)
	if err != nil {
		return nil, fmt.Errorf("invalid hex key: %w", err)
	}
	if len(key) != KeyLen {
		return nil, fmt.Errorf("key must be %d bytes", KeyLen)
	}
	return key, nil
}

func Encrypt(data []byte, passphrase string) ([]byte, error) {
	salt := make([]byte, SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to read salt: %w", err)
	}

	key := DeriveKey(passphrase, salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, NonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to read nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)

	enc := encryptedFile{
		Version: Version,
		Salt:    salt,
		Nonce:   nonce,
		Cipher:  ciphertext,
	}

	return encodeEncryptedFile(enc), nil
}

func EncryptWithKey(data []byte, key []byte) ([]byte, error) {
	salt := make([]byte, SaltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, fmt.Errorf("failed to read salt: %w", err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	nonce := make([]byte, NonceLen)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("failed to read nonce: %w", err)
	}

	ciphertext := gcm.Seal(nil, nonce, data, nil)

	enc := encryptedFile{
		Version: Version,
		Salt:    salt,
		Nonce:   nonce,
		Cipher:  ciphertext,
	}

	return encodeEncryptedFile(enc), nil
}

func encodeEncryptedFile(enc encryptedFile) []byte {
	result := []byte{enc.Version}
	result = append(result, enc.Salt...)
	result = append(result, enc.Nonce...)
	result = append(result, enc.Cipher...)
	return []byte(base64.StdEncoding.EncodeToString(result))
}

func decodeEncryptedFile(data []byte) (*encryptedFile, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	if len(decoded) < 1+SaltLen+NonceLen {
		return nil, errors.New("invalid encrypted file format")
	}

	enc := &encryptedFile{
		Version: decoded[0],
		Salt:    decoded[1 : 1+SaltLen],
		Nonce:   decoded[1+SaltLen : 1+SaltLen+NonceLen],
		Cipher:  decoded[1+SaltLen+NonceLen:],
	}

	return enc, nil
}

func Decrypt(data []byte, passphrase string) ([]byte, error) {
	enc, err := decodeEncryptedFile(data)
	if err != nil {
		return nil, err
	}

	if enc.Version != Version {
		return nil, ErrInvalidVersion
	}

	key := DeriveKey(passphrase, enc.Salt)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, enc.Nonce, enc.Cipher, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecryption, err)
	}

	return plaintext, nil
}

func DecryptWithKey(data []byte, key []byte) ([]byte, error) {
	enc, err := decodeEncryptedFile(data)
	if err != nil {
		return nil, err
	}

	if enc.Version != Version {
		return nil, ErrInvalidVersion
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	plaintext, err := gcm.Open(nil, enc.Nonce, enc.Cipher, nil)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrDecryption, err)
	}

	return plaintext, nil
}

func Verify(data []byte, passphrase string) error {
	_, err := Decrypt(data, passphrase)
	return err
}

func VerifyWithKey(data []byte, key []byte) error {
	_, err := DecryptWithKey(data, key)
	return err
}

func EncryptFile(inputPath, outputPath, passphrase string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	encrypted, err := Encrypt(data, passphrase)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	err = os.WriteFile(outputPath, encrypted, 0600)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func DecryptFile(inputPath, outputPath, passphrase string) error {
	data, err := os.ReadFile(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}

	decrypted, err := Decrypt(data, passphrase)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	err = os.WriteFile(outputPath, decrypted, 0600)
	if err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}

func HashPassword(password string) string {
	hash := sha256.Sum256([]byte(password))
	return hex.EncodeToString(hash[:])
}
