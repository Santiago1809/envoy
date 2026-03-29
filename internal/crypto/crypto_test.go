package crypto

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateKey(t *testing.T) {
	key1, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	if len(key1) != 64 {
		t.Errorf("key length = %d, want 64", len(key1))
	}

	key2, err := GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey() error = %v", err)
	}

	if key1 == key2 {
		t.Error("GenerateKey() should generate unique keys")
	}
}

func TestEncryptDecrypt(t *testing.T) {
	plaintext := []byte("Hello, World! This is a secret message.")
	passphrase := "my-secret-passphrase"

	encrypted, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatalf("Encrypt() error = %v", err)
	}

	if len(encrypted) == 0 {
		t.Error("encrypted data should not be empty")
	}

	decrypted, err := Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatalf("Decrypt() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted = %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestEncryptDecryptWithKey(t *testing.T) {
	plaintext := []byte("Hello, World! This is a secret message.")

	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	keyBytes, err := DeriveKeyFromHex(key)
	if err != nil {
		t.Fatal(err)
	}

	encrypted, err := EncryptWithKey(plaintext, keyBytes)
	if err != nil {
		t.Fatalf("EncryptWithKey() error = %v", err)
	}

	decrypted, err := DecryptWithKey(encrypted, keyBytes)
	if err != nil {
		t.Fatalf("DecryptWithKey() error = %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("decrypted = %q, want %q", string(decrypted), string(plaintext))
	}
}

func TestDecryptWrongPassphrase(t *testing.T) {
	plaintext := []byte("Secret message")
	passphrase := "correct-passphrase"

	encrypted, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	_, err = Decrypt(encrypted, "wrong-passphrase")
	if err == nil {
		t.Error("Decrypt() with wrong passphrase should fail")
	}
}

func TestVerify(t *testing.T) {
	plaintext := []byte("Secret message")
	passphrase := "my-passphrase"

	encrypted, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	err = Verify(encrypted, passphrase)
	if err != nil {
		t.Errorf("Verify() error = %v", err)
	}

	err = Verify(encrypted, "wrong-passphrase")
	if err == nil {
		t.Error("Verify() with wrong passphrase should fail")
	}
}

func TestEncryptFile(t *testing.T) {
	dir := t.TempDir()

	inputFile := filepath.Join(dir, "input.env")
	outputFile := filepath.Join(dir, "input.env.enc")

	content := "SECRET=value\nAPI_KEY=12345"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	err := EncryptFile(inputFile, outputFile, "password")
	if err != nil {
		t.Fatalf("EncryptFile() error = %v", err)
	}

	encryptedData, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatal(err)
	}

	if len(encryptedData) == 0 {
		t.Error("encrypted file should not be empty")
	}
}

func TestDecryptFile(t *testing.T) {
	dir := t.TempDir()

	inputFile := filepath.Join(dir, "input.env")
	encryptedFile := filepath.Join(dir, "input.env.enc")
	decryptedFile := filepath.Join(dir, "decrypted.env")

	content := "SECRET=value\nAPI_KEY=12345"
	if err := os.WriteFile(inputFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	passphrase := "password"

	err := EncryptFile(inputFile, encryptedFile, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	err = DecryptFile(encryptedFile, decryptedFile, passphrase)
	if err != nil {
		t.Fatalf("DecryptFile() error = %v", err)
	}

	decryptedContent, err := os.ReadFile(decryptedFile)
	if err != nil {
		t.Fatal(err)
	}

	if string(decryptedContent) != content {
		t.Errorf("decrypted content = %q, want %q", string(decryptedContent), content)
	}
}

func TestEncryptEmptyFile(t *testing.T) {
	plaintext := []byte("")
	passphrase := "password"

	encrypted, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	if string(decrypted) != "" {
		t.Error("empty file should decrypt to empty")
	}
}

func TestEncryptLargeFile(t *testing.T) {
	plaintext := make([]byte, 1024*1024)
	for i := range plaintext {
		plaintext[i] = byte(i % 256)
	}

	passphrase := "password"

	encrypted, err := Encrypt(plaintext, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	decrypted, err := Decrypt(encrypted, passphrase)
	if err != nil {
		t.Fatal(err)
	}

	if string(decrypted) != string(plaintext) {
		t.Error("large file encryption/decryption failed")
	}
}

func TestDeriveKeyFromHex(t *testing.T) {
	key, err := GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	keyBytes, err := DeriveKeyFromHex(key)
	if err != nil {
		t.Fatalf("DeriveKeyFromHex() error = %v", err)
	}

	if len(keyBytes) != KeyLen {
		t.Errorf("key length = %d, want %d", len(keyBytes), KeyLen)
	}

	_, err = DeriveKeyFromHex("invalid")
	if err == nil {
		t.Error("DeriveKeyFromHex() should fail for invalid hex")
	}

	_, err = DeriveKeyFromHex("0000000000000000000000000000000000000000000000000000000000000000")
	if err != nil {
		t.Error("DeriveKeyFromHex() should work for valid 64-char hex")
	}
}

func TestHashPassword(t *testing.T) {
	hash1 := HashPassword("password")
	hash2 := HashPassword("password")
	hash3 := HashPassword("different")

	if hash1 != hash2 {
		t.Error("same password should produce same hash")
	}

	if hash1 == hash3 {
		t.Error("different passwords should produce different hashes")
	}
}
