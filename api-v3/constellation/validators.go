package v3constellation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	// Key for the error code when an exit message already exists for the pubkey being submitted
	ExitMessageExistsKey string = "exit_message_exists"

	// Route for interacting with the list of validators
	ValidatorsPath string = "validators"
)

var (
	// Exit message already exists for the pubkey being submitted
	ErrExitMessageExists error = errors.New("exit message already exists for the pubkey")
)

// Validator status info
type ValidatorStatus struct {
	Pubkey              beacon.ValidatorPubkey `json:"pubkey"`
	RequiresExitMessage bool                   `json:"requiresExitMessage"`
}

// Response to a validators request
type ValidatorsData struct {
	Validators []ValidatorStatus `json:"validators"`
}

// Details of an exit message
type ExitMessageDetails struct {
	Epoch          string `json:"epoch"`
	ValidatorIndex string `json:"validatorIndex"`
}

// Voluntary exit message
type ExitMessage struct {
	Message   ExitMessageDetails `json:"message"`
	Signature string             `json:"signature"`
}

// Data for a pubkey's voluntary exit message
type ExitData struct {
	Pubkey      string      `json:"pubkey"`
	ExitMessage ExitMessage `json:"exitMessage"`
}

// Data for a pubkey's encrypted voluntary exit message
type EncryptedExitData struct {
	Pubkey      string `json:"pubkey"`
	ExitMessage string `json:"exitMessage"`
}

// Request body for submitting exit data
type Validators_PatchBody struct {
	ExitData []EncryptedExitData `json:"exitData"`
}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node on the provided deployment and vault
func (c *V3ConstellationClient) Validators_Get(ctx context.Context, logger *slog.Logger, deployment string) (ValidatorsData, error) {
	// Send the request
	path := ConstellationPrefix + deployment + "/" + ValidatorsPath
	code, response, err := common.SubmitRequest[ValidatorsData](c.commonClient, ctx, logger, true, http.MethodGet, nil, nil, path)
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

		case MissingWhitelistedNodeAddressKey:
			// Node address not whitelisted for deployment
			return ValidatorsData{}, ErrMissingWhitelistedNodeAddress

		case IncorrectNodeAddressKey:
			// Incorrect node address
			return ValidatorsData{}, ErrIncorrectNodeAddress
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return ValidatorsData{}, common.ErrInvalidSession
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return ValidatorsData{}, common.ErrInvalidPermissions
		}
	}
	return ValidatorsData{}, fmt.Errorf("nodeset server responded to validators-get request with code %d: [%s]", code, response.Message)
}

// Submit signed exit data to NodeSet
// func (c *V3ConstellationClient) Validators_Patch(ctx context.Context, logger *slog.Logger, deployment string, exitData []common.EncryptedExitData) error {
// 	// Create the request body
// 	body := Validators_PatchBody{
// 		ExitData: make([]EncryptedExitData, len(exitData)),
// 	}
// 	for i, data := range exitData {
// 		body.ExitData[i] = EncryptedExitData{
// 			Pubkey:      data.Pubkey,
// 			ExitMessage: data.ExitMessage,
// 		}
// 	}
// 	jsonData, err := json.Marshal(body)
// 	if err != nil {
// 		return fmt.Errorf("error marshalling exit data to JSON: %w", err)
// 	}
// 	common.SafeDebugLog(logger, "Prepared validators PATCH body",
// 		"body", body,
// 	)

// 	// Send the request
// 	path := ConstellationPrefix + deployment + "/" + ValidatorsPath
// 	code, response, err := common.SubmitRequest[struct{}](c.commonClient, ctx, logger, true, http.MethodPatch, bytes.NewBuffer(jsonData), nil, path)
// 	if err != nil {
// 		return fmt.Errorf("error submitting exit data: %w", err)
// 	}

// 	// Handle response based on return code
// 	switch code {
// 	case http.StatusOK:
// 		return nil

// 	case http.StatusBadRequest:
// 		switch response.Error {
// 		case common.InvalidDeploymentKey:
// 			// Invalid deployment
// 			return common.ErrInvalidDeployment

// 		case common.MalformedInputKey:
// 			// Invalid input
// 			return common.ErrMalformedInput

// 		case common.InvalidValidatorOwnerKey:
// 			// Invalid validator owner
// 			return common.ErrInvalidValidatorOwner

// 		case common.InvalidExitMessageKey:
// 			// Invalid exit message
// 			return common.ErrInvalidExitMessage

// 		case MissingWhitelistedNodeAddressKey:
// 			// Node address not whitelisted for deployment
// 			return ErrMissingWhitelistedNodeAddress

// 		case IncorrectNodeAddressKey:
// 			// Incorrect node address
// 			return ErrIncorrectNodeAddress

// 		case ExitMessageExistsKey:
// 			// Exit message already exists for the pubkey being submitted
// 			return ErrExitMessageExists
// 		}

// 	case http.StatusUnauthorized:
// 		switch response.Error {
// 		case common.InvalidSessionKey:
// 			// Invalid or expired session
// 			return common.ErrInvalidSession
// 		}

// 	case http.StatusForbidden:
// 		switch response.Error {
// 		case common.InvalidPermissionsKey:
// 			// The user doesn't have permission to do this
// 			return common.ErrInvalidPermissions
// 		}
// 	}
// 	return fmt.Errorf("nodeset server responded to validators-patch request with code %d: [%s]", code, response.Message)
// }
