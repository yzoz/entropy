package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters used for deriving the key-encryption key (KEK) from a
// passphrase. They follow the RFC 9106 "second recommended option" and are
// stored alongside the ciphertext so parameters can evolve without breaking
// existing key files.
const (
	argonTime    = 1
	argonMemory  = 64 * 1024 // 64 MiB
	argonThreads = 4
	argonKeyLen  = 32
	saltLen      = 16
)

// ErrWrongPassphrase is returned when decryption fails, which almost always
// means an incorrect passphrase (AES-GCM also authenticates the ciphertext).
var ErrWrongPassphrase = errors.New("wrong passphrase or corrupted key file")

// EncryptedKey is the on-disk JSON representation of a passphrase-protected
// private key.
type EncryptedKey struct {
	Version int    `json:"version"`
	KDF     string `json:"kdf"` // "argon2id"
	Cipher  string `json:"cipher"`
	Time    uint32 `json:"time"`
	Memory  uint32 `json:"memory"`
	Threads uint8  `json:"threads"`
	Salt    string `json:"salt"` // base64
	Nonce   string `json:"nonce"`
	Data    string `json:"data"` // base64 ciphertext
	Address string `json:"address"`
	Public  string `json:"public"` // hex, convenient for import/export
}

// EncryptPrivateKey wraps a hex-encoded Ed25519 private key with a passphrase.
func EncryptPrivateKey(privHex, passphrase string, pubHex string) ([]byte, error) {
	if passphrase == "" {
		return nil, errors.New("empty passphrase")
	}
	salt := make([]byte, saltLen)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return nil, err
	}
	kek := argon2.IDKey([]byte(passphrase), salt, argonTime, argonMemory, argonThreads, argonKeyLen)
	block, err := aes.NewCipher(kek)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	ct := gcm.Seal(nil, nonce, []byte(privHex), nil)

	pub, _ := DecodePublic(pubHex)
	out := EncryptedKey{
		Version: 1,
		KDF:     "argon2id",
		Cipher:  "aes-256-gcm",
		Time:    argonTime,
		Memory:  argonMemory,
		Threads: argonThreads,
		Salt:    b64(salt),
		Nonce:   b64(nonce),
		Data:    b64(ct),
		Address: Address(pub),
		Public:  pubHex,
	}
	return json.MarshalIndent(out, "", "  ")
}

// DecryptPrivateKey decrypts a key file produced by EncryptPrivateKey.
func DecryptPrivateKey(raw []byte, passphrase string) (privHex, pubHex string, err error) {
	var k EncryptedKey
	if err := json.Unmarshal(raw, &k); err != nil {
		return "", "", fmt.Errorf("unreadable key file: %w", err)
	}
	if k.KDF != "argon2id" || k.Cipher != "aes-256-gcm" {
		return "", "", fmt.Errorf("unsupported key format: %s/%s", k.KDF, k.Cipher)
	}
	salt, err := unb64(k.Salt)
	if err != nil {
		return "", "", err
	}
	nonce, err := unb64(k.Nonce)
	if err != nil {
		return "", "", err
	}
	ct, err := unb64(k.Data)
	if err != nil {
		return "", "", err
	}
	kek := argon2.IDKey([]byte(passphrase), salt, k.Time, k.Memory, k.Threads, argonKeyLen)
	block, err := aes.NewCipher(kek)
	if err != nil {
		return "", "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}
	pt, err := gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", "", ErrWrongPassphrase
	}
	return string(pt), k.Public, nil
}
