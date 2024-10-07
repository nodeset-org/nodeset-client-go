package common

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"io"

	"filippo.io/age"
	"github.com/goccy/go-json"
)

// Encrypt a hex-encoded signed exit message for submission to NodeSet.io using the age encryption library.
// The recipient is the public key of the recipient that will decrypt the message, typically the NodeSet.io service.
func EncryptSignedExitMessage(message ExitMessage, recipientPubkey string) (string, error) {
	// Serialize the message to JSON
	bytes, err := json.Marshal(message)
	if err != nil {
		return "", fmt.Errorf("error serializing exit message: %w", err)
	}

	// Encrypt the message
	encrypted, err := EncryptMessage(string(bytes), recipientPubkey)
	if err != nil {
		return "", fmt.Errorf("error encrypting exit message: %w", err)
	}

	// Encode the encrypted message as hex
	encoded := hex.EncodeToString(encrypted)
	return encoded, nil
}

// Encrypt an arbitrary message for submission to NodeSet.io using the age encryption library.
// The recipient is the public key of the recipient that will decrypt the message, typically the NodeSet.io service.
func EncryptMessage(message string, recipientPubkey string) ([]byte, error) {
	recipient, err := age.ParseX25519Recipient(recipientPubkey)
	if err != nil {
		return nil, fmt.Errorf("error parsing pubkey: %w", err)
	}

	// Encrypt the message
	out := &bytes.Buffer{}
	writer, err := age.Encrypt(out, recipient)
	if err != nil {
		return nil, fmt.Errorf("error creating encryption writer: %w", err)
	}
	_, err = io.WriteString(writer, message)
	if err != nil {
		return nil, fmt.Errorf("error writing message to encryption buffer: %w", err)
	}
	err = writer.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing encryption writer: %w", err)
	}
	return out.Bytes(), nil
}
