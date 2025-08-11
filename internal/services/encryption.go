package services

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// EncryptionService maneja la encriptación y desencriptación de datos sensibles
type EncryptionService struct {
	key []byte
}

// NewEncryptionService crea una nueva instancia del servicio de encriptación
func NewEncryptionService(key string) (*EncryptionService, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("encryption key must be exactly 32 bytes")
	}

	return &EncryptionService{
		key: []byte(key),
	}, nil
}

// Encrypt encripta un texto plano
func (s *EncryptionService) Encrypt(plaintext string) (string, error) {
	// Crear cipher block
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Crear GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Crear nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to create nonce: %w", err)
	}

	// Encriptar
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)

	// Codificar en base64
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt desencripta un texto encriptado
func (s *EncryptionService) Decrypt(encryptedText string) (string, error) {
	// Decodificar de base64
	ciphertext, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", fmt.Errorf("failed to decode base64: %w", err)
	}

	// Crear cipher block
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", fmt.Errorf("failed to create cipher block: %w", err)
	}

	// Crear GCM
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("failed to create GCM: %w", err)
	}

	// Extraer nonce
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]

	// Desencriptar
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}

	return string(plaintext), nil
}

// EncryptAccessToken encripta un access token
func (s *EncryptionService) EncryptAccessToken(token string) (string, error) {
	if token == "" {
		return "", fmt.Errorf("token cannot be empty")
	}

	return s.Encrypt(token)
}

// DecryptAccessToken desencripta un access token
func (s *EncryptionService) DecryptAccessToken(encryptedToken string) (string, error) {
	if encryptedToken == "" {
		return "", fmt.Errorf("encrypted token cannot be empty")
	}

	return s.Decrypt(encryptedToken)
}

// IsEncrypted verifica si un texto está encriptado
func (s *EncryptionService) IsEncrypted(text string) bool {
	if text == "" {
		return false
	}

	// Intentar decodificar de base64
	_, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return false
	}

	// Verificar que tenga el tamaño mínimo para un texto encriptado
	// (nonce + al menos algunos bytes de datos)
	return len(text) > 50
}
