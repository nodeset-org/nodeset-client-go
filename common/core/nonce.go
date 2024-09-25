package core

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common"
)

const (
	// Route for getting a nonce from the NodeSet server
	NoncePath string = "nonce"
)

// Data used returned from nonce requests
type NonceData struct {
	// The nonce for the session request
	Nonce string `json:"nonce"`

	// The auth token for the session if approved
	Token string `json:"token"`
}

// Get a nonce from the NodeSet server for a new session
func Nonce(c *common.CommonNodeSetClient, ctx context.Context, logger *slog.Logger, noncePath string) (NonceData, error) {
	// Get the nonce
	code, nonceResponse, err := common.SubmitRequest[NonceData](c, ctx, logger, false, http.MethodGet, nil, nil, noncePath)
	if err != nil {
		return NonceData{}, fmt.Errorf("error getting nonce: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return nonceResponse.Data, nil
	}
	return NonceData{}, fmt.Errorf("nodeset server responded to nonce request with code %d: [%s]", code, nonceResponse.Message)
}
