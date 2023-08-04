package crypt

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
)

func GeneratePrivateKeyPEM() (string, error) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return "", fmt.Errorf("failed to generate RSA private key: %w", err)
	}

	return ConvertPrivateKeyToPEM(privateKey)
}

func GeneratePuublicKeyPEM(privateKey string) (string, error) {
	block, _ := pem.Decode([]byte(privateKey))

	parsedPrivateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return "", fmt.Errorf("failed to parse PKCS8 private key: %w", err)
	}

	rsaPrivateKey, ok := parsedPrivateKey.(*rsa.PrivateKey)
	if !ok {
		return "", fmt.Errorf("unsupported private key type: %T", parsedPrivateKey)
	}

	publickKey := &rsaPrivateKey.PublicKey
	derRsaPublicKey, err := x509.MarshalPKIXPublicKey(publickKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal RSA public key: %w", err)
	}

	pubKeyBlock := &pem.Block{
		Type:  "PUBLIC KEY",
		Bytes: derRsaPublicKey,
	}

	pubBuf := bytes.NewBufferString("")
	if err = pem.Encode(pubBuf, pubKeyBlock); err != nil {
		return "", fmt.Errorf("cannot encode RSA public key: %v", err)
	}

	return pubBuf.String(), nil
}

func ConvertPrivateKeyToPEM(privateKey crypto.PrivateKey) (string, error) {
	derRsaPrivateKey, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		return "", fmt.Errorf("failed to marshal RSA private key: %w", err)
	}
	priKeyBlock := &pem.Block{
		Type:  "PRIVATE KEY",
		Bytes: derRsaPrivateKey,
	}

	privBuf := bytes.NewBufferString("")
	if err = pem.Encode(privBuf, priKeyBlock); err != nil {
		return "", fmt.Errorf("cannot encode RSA private key: %v", err)
	}

	return privBuf.String(), nil
}

func ConvertPrivateKey(privateKeyPEM string) (crypto.PrivateKey, error) {
	pemBlock, _ := pem.Decode([]byte(privateKeyPEM))

	var err error
	var privateKey crypto.PrivateKey
	switch pemBlock.Type {
	case "PRIVATE KEY":
		privateKey, err = x509.ParsePKCS8PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	case "RSA PRIVATE KEY":
		privateKey, err = x509.ParsePKCS1PrivateKey(pemBlock.Bytes)
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported private key type: %s", pemBlock.Type)
	}

	return privateKey, nil
}
