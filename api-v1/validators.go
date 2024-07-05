package apiv1

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	// Route for interacting with the list of validators
	ValidatorsPath string = "validators"

	// The requester doesn't own the provided validator
	InvalidValidatorOwnerKey string = "invalid_validator_owner"

	// The exit message provided was invalid
	InvalidExitMessage string = "invalid_exit_message"
)

var (
	// The requester doesn't own the provided validator
	ErrInvalidValidatorOwner error = fmt.Errorf("this node doesn't own one of the provided validators")

	// The exit message provided was invalid
	ErrInvalidExitMessage error = fmt.Errorf("the provided exit message was invalid")
)

type StakeWiseStatus string

const (
	// DepositData hasn't been uploaded to NodeSet yet
	StakeWiseStatus_Unknown StakeWiseStatus = "UNKNOWN"

	// DepositData uploaded to NodeSet, but hasn't been made part of a deposit data set yet
	StakeWiseStatus_Pending StakeWiseStatus = "PENDING"

	// DepositData uploaded to NodeSet, uploaded to StakeWise, but hasn't been activated on Beacon yet
	StakeWiseStatus_Uploaded StakeWiseStatus = "UPLOADED"

	// DepositData uploaded to NodeSet, uploaded to StakeWise, and the validator is active on Beacon
	StakeWiseStatus_Registered StakeWiseStatus = "REGISTERED"

	// DepositData uploaded to NodeSet, uploaded to StakeWise, and the validator is exited on Beacon
	StakeWiseStatus_Removed StakeWiseStatus = "REMOVED"
)

// Validator status info
type ValidatorStatus struct {
	Pubkey              beacon.ValidatorPubkey `json:"pubkey"`
	Status              StakeWiseStatus        `json:"status"`
	ExitMessageUploaded bool                   `json:"exitMessage"`
}

// Response to a validators request
type ValidatorsData struct {
	Validators []ValidatorStatus `json:"validators"`
}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node
func (c *NodeSetClient) Validators_Get(ctx context.Context, network string) (ValidatorsData, error) {
	// Create the request params
	queryParams := map[string]string{
		"network": network,
	}

	// Send the request
	code, response, err := SubmitRequest[ValidatorsData](c, ctx, true, http.MethodGet, nil, queryParams, c.routes.Validators)
	if err != nil {
		return ValidatorsData{}, fmt.Errorf("error getting registered validators: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case InvalidNetworkKey:
			// Network not known
			return ValidatorsData{}, ErrInvalidNetwork
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case InvalidSessionKey:
			// Invalid or expird session
			return ValidatorsData{}, ErrInvalidSession
		}
	}
	return ValidatorsData{}, fmt.Errorf("nodeset server responded to validators-get request with code %d: [%s]", code, response.Message)
}

// Details of an exit message
type ExitMessageDetails struct {
	Epoch          string `json:"epoch"`
	ValidatorIndex string `json:"validator_index"`
}

// Voluntary exit message
type ExitMessage struct {
	Message   ExitMessageDetails `json:"message"`
	Signature string             `json:"signature"`
}

// Data for a pubkey's voluntary exit message
type ExitData struct {
	Pubkey      string      `json:"pubkey"`
	ExitMessage ExitMessage `json:"exit_message"`
}

// Submit signed exit data to Nodeset
func (c *NodeSetClient) Validators_Patch(ctx context.Context, exitData []ExitData, network string) error {
	// Create the request body
	jsonData, err := json.Marshal(exitData)
	if err != nil {
		return fmt.Errorf("error marshalling exit data to JSON: %w", err)
	}

	// Create the request params
	params := map[string]string{
		"network": network,
	}

	// Submit the request
	code, response, err := SubmitRequest[struct{}](c, ctx, true, http.MethodPatch, bytes.NewBuffer(jsonData), params, c.routes.Validators)
	if err != nil {
		return fmt.Errorf("error submitting exit data: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return nil

	case http.StatusBadRequest:
		switch response.Error {
		case InvalidNetworkKey:
			// Network not known
			return ErrInvalidNetwork

		case MalformedInputKey:
			// Invalid input
			return ErrMalformedInput

		case InvalidValidatorOwnerKey:
			// Invalid validator owner
			return ErrInvalidValidatorOwner

		case InvalidExitMessage:
			// Invalid exit message
			return ErrInvalidExitMessage
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case InvalidSessionKey:
			// Invalid or expird session
			return ErrInvalidSession
		}
	}
	return fmt.Errorf("nodeset server responded to validators-patch request with code %d: [%s]", code, response.Message)
}
