package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/utils"
)

const (
	// Format for signing login messages
	LoginMessageFormat string = `{"nonce":"%s","address":"%s"}`

	// Route for logging into the NodeSet server
	LoginPath string = "login"

	// The provided nonce didn't match an expected one
	InvalidNonceKey string = "invalid_nonce"

	// Value of the auth response header if the node hasn't registered yet
	UnregisteredAddressKey string = "unregistered_address"
)

var (
	// The provided nonce didn't match an expected one
	ErrInvalidNonce error = errors.New("invalid nonce provided for login")

	// The node hasn't been registered with the NodeSet server yet
	ErrUnregisteredNode error = errors.New("node hasn't been registered with the NodeSet server yet")
)

// Request to log into the NodeSet server
type LoginRequest struct {
	// The nonce for the session request
	Nonce string `json:"nonce"`

	// The node's wallet address
	Address string `json:"address"`

	// Signature of the login request
	Signature string `json:"signature"` // Must be 0x-prefixed hex encoded
}

// Response to a login request
type LoginData struct {
	// The auth token for the session if approved
	Token string `json:"token"`
}

// Logs into the NodeSet server, starting a new session
func Login(c *common.CommonNodeSetClient, ctx context.Context, logger *slog.Logger, nonce string, address ethcommon.Address, signature []byte, loginPath string) (LoginData, error) {
	// Create the request body
	addressString := address.Hex()
	signatureString := utils.EncodeHexWithPrefix(signature)
	request := LoginRequest{
		Nonce:     nonce,
		Address:   addressString,
		Signature: signatureString,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return LoginData{}, fmt.Errorf("error marshalling login request: %w", err)
	}
	common.SafeDebugLog(logger, "Prepared login body",
		"body", request,
	)

	// Submit the request
	code, response, err := common.SubmitRequest[LoginData](c, ctx, logger, true, http.MethodPost, bytes.NewBuffer(jsonData), nil, loginPath)
	if err != nil {
		return LoginData{}, fmt.Errorf("error submitting login request: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Login successful, session established
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidSignatureKey:
			// Invalid signature
			return LoginData{}, common.ErrInvalidSignature

		case common.MalformedInputKey:
			// Malformed input
			return LoginData{}, common.ErrMalformedInput

		case InvalidNonceKey:
			// Invalid nonce
			return LoginData{}, ErrInvalidNonce
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case UnregisteredAddressKey:
			// Node hasn't been registered yet
			return LoginData{}, ErrUnregisteredNode

		case common.InvalidSessionKey:
			// The nonce wasn't expected?
			return LoginData{}, common.ErrInvalidSession
		}
	}
	return LoginData{}, fmt.Errorf("nodeset server responded to login request with code %d: [%s]", code, response.Message)
}
