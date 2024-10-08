package auth

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/nodeset-org/nodeset-client-go/common/core"
	nsutil "github.com/nodeset-org/nodeset-client-go/utils"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

const (
	// Header used for the wallet signature during a deposit data upload
	authHeader string = "Authorization"

	// Format for the authorization header
	authHeaderFormat string = "Bearer %s"
)

var (
	ErrAuthHeader        error = errors.New("invalid auth header")
	ErrMissingAuthHeader error = errors.New("missing auth header")
)

// Creates a signature for node registration
func GetSignatureForRegistration(email string, nodeAddress common.Address, privateKey *ecdsa.PrivateKey, nodeRegistrationMessageFormat string) ([]byte, error) {
	message := fmt.Sprintf(nodeRegistrationMessageFormat, email, nodeAddress.Hex())
	return nsutil.CreateSignature([]byte(message), privateKey)
}

// Creates a signature for node registration
func GetSignatureForLogin(nonce string, nodeAddress common.Address, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	message := fmt.Sprintf(core.LoginMessageFormat, nonce, nodeAddress.Hex())
	return nsutil.CreateSignature([]byte(message), privateKey)
}

// Verifies a signature for node registration
func VerifyRegistrationSignature(email string, nodeAddress common.Address, signature []byte, nodeRegistrationMessageFormat string) error {
	message := fmt.Sprintf(nodeRegistrationMessageFormat, email, nodeAddress.Hex())
	address, err := getAddressFromSignature([]byte(message), signature)
	if err != nil {
		return fmt.Errorf("error verifying signature: %w", err)
	}
	if address != nodeAddress {
		return errors.New("signature does not match node address")
	}
	return nil
}

// Verifies a signature for logging in
func VerifyLoginSignature(nonce string, nodeAddress common.Address, signature []byte) error {
	message := fmt.Sprintf(core.LoginMessageFormat, nonce, nodeAddress.Hex())
	address, err := getAddressFromSignature([]byte(message), signature)
	if err != nil {
		return fmt.Errorf("error verifying signature: %w", err)
	}
	if address != nodeAddress {
		return errors.New("signature does not match node address")
	}
	return nil
}

// Gets the session token from a request
func GetSessionTokenFromRequest(r *http.Request) (string, error) {
	// Get the auth header
	authHeaderVals, exists := r.Header[authHeader]
	if !exists || len(authHeaderVals) == 0 {
		return "", ErrMissingAuthHeader
	}
	authHeaderVal := authHeaderVals[0]
	if !strings.HasPrefix(authHeaderVal, "Bearer ") {
		return "", ErrAuthHeader
	}

	// Get the session token
	elements := strings.Split(authHeaderVal, " ")
	if len(elements) != 2 {
		return "", ErrAuthHeader
	}
	return elements[1], nil
}

// Adds an authorization header to an HTTP request
func AddAuthorizationHeader(request *http.Request, sessionToken string) {
	request.Header.Set(authHeader, fmt.Sprintf(authHeaderFormat, sessionToken))
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")
}

// Gets the address of the private key used to sign a message from a signature
func getAddressFromSignature(message []byte, signature []byte) (common.Address, error) {
	// Fix the ECDSA 'v' (see https://medium.com/mycrypto/the-magic-of-digital-signatures-on-ethereum-98fe184dc9c7#:~:text=The%20version%20number,2%E2%80%9D%20was%20introduced)
	if signature[crypto.RecoveryIDOffset] >= 4 {
		signature[crypto.RecoveryIDOffset] -= 27
	}

	// Get the address
	messageHash := accounts.TextHash(message)
	pubkeyBytes, err := crypto.SigToPub(messageHash, signature)
	if err != nil {
		return common.Address{}, fmt.Errorf("error recovering pubkey from signature: %w", err)
	}

	return crypto.PubkeyToAddress(*pubkeyBytes), nil
}
