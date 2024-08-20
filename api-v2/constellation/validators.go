package v2constellation

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	// Route for interacting with the list of validators
	ValidatorsPath string = "validators"
)

// Validator status info
type ValidatorStatus struct {
	Pubkey              beacon.ValidatorPubkey `json:"pubkey"`
	ExitMessageUploaded bool                   `json:"exitMessage"`
}

// Response to a validators request
type ValidatorsData struct {
	Validators []ValidatorStatus `json:"validators"`
}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node on the provided deployment and vault
func (c *V2ConstellationClient) Validators_Get(ctx context.Context, deployment string) (ValidatorsData, error) {
	// Send the request
	path := ConstellationPrefix + deployment + "/" + ValidatorsPath
	code, response, err := common.SubmitRequest[ValidatorsData](c.commonClient, ctx, true, http.MethodGet, nil, nil, path)
	if err != nil {
		return ValidatorsData{}, fmt.Errorf("error getting registered validators: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return ValidatorsData{}, common.ErrInvalidDeployment
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expird session
			return ValidatorsData{}, common.ErrInvalidSession
		}
	}
	return ValidatorsData{}, fmt.Errorf("nodeset server responded to validators-get request with code %d: [%s]", code, response.Message)
}

// Submit signed exit data to NodeSet
func (c *V2ConstellationClient) Validators_Patch(ctx context.Context, deployment string, vault ethcommon.Address, exitData []common.ExitData) error {
	// Create the request body
	jsonData, err := json.Marshal(exitData)
	if err != nil {
		return fmt.Errorf("error marshalling exit data to JSON: %w", err)
	}

	// Send the request
	path := ConstellationPrefix + deployment + "/" + vault.Hex() + "/" + ValidatorsPath
	code, response, err := common.SubmitRequest[struct{}](c.commonClient, ctx, true, http.MethodPatch, bytes.NewBuffer(jsonData), nil, path)
	if err != nil {
		return fmt.Errorf("error submitting exit data: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return common.ErrInvalidDeployment

		case common.MalformedInputKey:
			// Invalid input
			return common.ErrMalformedInput

		case common.InvalidValidatorOwnerKey:
			// Invalid validator owner
			return common.ErrInvalidValidatorOwner

		case common.InvalidExitMessage:
			// Invalid exit message
			return common.ErrInvalidExitMessage
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expird session
			return common.ErrInvalidSession
		}
	}
	return fmt.Errorf("nodeset server responded to validators-patch request with code %d: [%s]", code, response.Message)
}
