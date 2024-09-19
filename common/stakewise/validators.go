package stakewise

import (
	"bytes"
	"context"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	// Route for interacting with the list of validators
	ValidatorsPath string = "validators"
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
func Validators_Get(c *common.CommonNodeSetClient, ctx context.Context, params map[string]string, validatorsPath string) (int, *common.NodeSetResponse[ValidatorsData], error) {
	// Send the request
	code, response, err := common.SubmitRequest[ValidatorsData](c, ctx, true, http.MethodGet, nil, params, validatorsPath)
	if err != nil {
		return code, nil, fmt.Errorf("error getting registered validators: %w", err)
	}

	// Handle common errors
	switch code {
	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expird session
			return code, nil, common.ErrInvalidSession
		}
	}
	return code, &response, nil
}

// Submit signed exit data to Nodeset
func Validators_Patch(c *common.CommonNodeSetClient, ctx context.Context, exitData any, params map[string]string, validatorsPath string) (int, *common.NodeSetResponse[struct{}], error) {
	// Create the request body
	jsonData, err := json.Marshal(exitData)
	if err != nil {
		return -1, nil, fmt.Errorf("error marshalling exit data to JSON: %w", err)
	}

	// Submit the request
	code, response, err := common.SubmitRequest[struct{}](c, ctx, true, http.MethodPatch, bytes.NewBuffer(jsonData), params, validatorsPath)
	if err != nil {
		return code, nil, fmt.Errorf("error submitting exit data: %w", err)
	}

	// Handle common errors
	switch code {
	case http.StatusBadRequest:
		switch response.Error {
		case common.MalformedInputKey:
			// Invalid input
			return code, nil, common.ErrMalformedInput

		case common.InvalidValidatorOwnerKey:
			// Invalid validator owner
			return code, nil, common.ErrInvalidValidatorOwner

		case common.InvalidExitMessage:
			// Invalid exit message
			return code, nil, common.ErrInvalidExitMessage
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expird session
			return code, nil, common.ErrInvalidSession
		}
	}
	return code, &response, nil
}
