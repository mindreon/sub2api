package voucher

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"
	"os"
	"strings"
)

func pinEncryptionKey() []byte {
	raw := strings.TrimSpace(os.Getenv("KVOUCHER_PIN_ENCRYPTION_KEY"))
	if raw == "" {
		raw = strings.TrimSpace(os.Getenv("JWT_SECRET"))
	}
	if raw == "" {
		raw = "voucher-pin-dev-key"
	}
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

// EncryptPIN stores a PIN for persistence.
func EncryptPIN(plain string) (string, error) {
	key := pinEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plain), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptPIN reads a stored PIN.
func DecryptPIN(encoded string) (string, error) {
	raw, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	key := pinEncryptionKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, ciphertext := raw[:gcm.NonceSize()], raw[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}

// MaskPIN shows only last segment for list views.
func MaskPIN(pin string) string {
	parts := strings.Split(pin, "-")
	if len(parts) == 0 {
		return "****"
	}
	last := parts[len(parts)-1]
	if len(last) <= 4 {
		return "****-" + last
	}
	return "****-****-****-" + last[len(last)-4:]
}
