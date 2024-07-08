package utils

import (
	"crypto/ecdsa"
	"fmt"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
)

func CreateSignature(message []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	messageHash := accounts.TextHash(message)
	signature, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		return nil, fmt.Errorf("error signing message: %w", err)
	}

	// Fix the ECDSA 'v' (see https://medium.com/mycrypto/the-magic-of-digital-signatures-on-ethereum-98fe184dc9c7#:~:text=The%20version%20number,2%E2%80%9D%20was%20introduced)
	signature[crypto.RecoveryIDOffset] += 27
	return signature, nil
}
