package welcomer

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"os"

	"github.com/gofrs/uuid"
)

// KeyType distinguishes between loading a public or private key
type KeyType string

const (
	PublicKey  KeyType = "PUBLIC"
	PrivateKey KeyType = "PRIVATE"
)

// LoadRSAKey loads an RSA public or private key from a file specified in environment variable
func LoadRSAKey(keyType KeyType) (interface{}, error) {
	var path string

	folder := os.Getenv("CUSTOM_BOT_KEY_FOLDER")
	if folder == "" {
		return nil, errors.New("CUSTOM_BOT_KEY_FOLDER not set")
	}

	switch keyType {
	case PublicKey:
		path = folder + "/public.pem"
	case PrivateKey:
		path = folder + "/private.pem"
	default:
		return nil, errors.New("invalid key type")
	}

	pemData, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read %s key file: %w", keyType, err)
	}

	block, _ := pem.Decode(pemData)
	if block == nil {
		return nil, errors.New("failed to decode PEM block")
	}

	switch keyType {
	case PublicKey:
		if block.Type != "PUBLIC KEY" {
			return nil, errors.New("expected PUBLIC KEY block")
		}

		pubKey, err := x509.ParsePKIXPublicKey(block.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}

		rsaPubKey, ok := pubKey.(*rsa.PublicKey)
		if !ok {
			return nil, errors.New("not an RSA public key")
		}

		return rsaPubKey, nil

	case PrivateKey:
		var privKey *rsa.PrivateKey

		switch block.Type {
		case "RSA PRIVATE KEY":
			privKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKCS1 private key: %w", err)
			}
		case "PRIVATE KEY":
			// For "PRIVATE KEY" block, we need to parse it as PKCS8
			privKeyInterface, err := x509.ParsePKCS8PrivateKey(block.Bytes)
			if err != nil {
				return nil, fmt.Errorf("failed to parse PKCS8 private key: %w", err)
			}

			var ok bool
			privKey, ok = privKeyInterface.(*rsa.PrivateKey)
			if !ok {
				return nil, errors.New("not an RSA private key")
			}
		default:
			return nil, fmt.Errorf("unexpected key type: %s", block.Type)
		}

		return privKey, nil
	}

	return nil, errors.New("unknown error")
}

func EncryptBotToken(token string, botID uuid.UUID) (string, error) {
	if token == "" {
		return "", nil
	}

	key, err := LoadRSAKey(PublicKey)
	if err != nil {
		return "", errors.New("failed to load public key: " + err.Error())
	}

	pubKey, ok := key.(*rsa.PublicKey)
	if !ok {
		return "", errors.New("invalid RSA public key")
	}

	cipherText, err := rsa.EncryptPKCS1v15(rand.Reader, pubKey, []byte(token))
	if err != nil {
		return "", errors.New("failed to encrypt token: " + err.Error())
	}

	return base64.StdEncoding.EncodeToString(cipherText), nil
}

func DecryptBotToken(encryptedToken string, botID uuid.UUID) (string, error) {
	if encryptedToken == "" {
		return "", nil
	}

	key, err := LoadRSAKey(PrivateKey)
	if err != nil {
		return "", err
	}

	privKey, ok := key.(*rsa.PrivateKey)
	if !ok {
		return "", errors.New("invalid RSA private key")
	}

	cipherBytes, err := base64.StdEncoding.DecodeString(encryptedToken)
	if err != nil {
		return "", errors.New("failed to decode base64 string: " + err.Error())
	}

	plainText, err := rsa.DecryptPKCS1v15(rand.Reader, privKey, cipherBytes)
	if err != nil {
		return "", errors.New("failed to decrypt token: " + err.Error())
	}

	return string(plainText), nil
}
