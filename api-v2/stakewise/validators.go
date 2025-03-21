package v2stakewise

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/rocket-pool/node-manager-core/beacon"
)

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

// Validator status info
type ValidatorStatus struct {
	Pubkey              beacon.ValidatorPubkey    `json:"pubkey"`
	Status              stakewise.StakeWiseStatus `json:"status"`
	ExitMessageUploaded bool                      `json:"exitMessage"`
}

// Response to a validators request
type ValidatorsData struct {
	Validators []ValidatorStatus `json:"validators"`
}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node on the provided deployment and vault
func (c *V2StakeWiseClient) Validators_Get(ctx context.Context, logger *slog.Logger, deployment string, vault ethcommon.Address) (ValidatorsData, error) {
	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.ValidatorsPath
	code, response, err := V2SubmitValidators_Get(c.commonClient, ctx, logger, nil, path)
	if err != nil {
		return ValidatorsData{}, err
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

		case stakewise.InvalidVaultKey:
			// Invalid vault
			return ValidatorsData{}, stakewise.ErrInvalidVault
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
func (c *V2StakeWiseClient) Validators_Patch(ctx context.Context, logger *slog.Logger, deployment string, vault ethcommon.Address, exitData []common.EncryptedExitData) error {
	// Create the request body
	body := Validators_PatchBody{
		ExitData: make([]EncryptedExitData, len(exitData)),
	}
	for i, data := range exitData {
		body.ExitData[i] = EncryptedExitData{
			Pubkey:      data.Pubkey,
			ExitMessage: data.ExitMessage,
		}
	}

	// Send the request
	common.SafeDebugLog(logger, "Prepared validators PATCH body",
		"body", body,
	)
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.ValidatorsPath
	code, response, err := stakewise.Validators_Patch(c.commonClient, ctx, logger, body, nil, path)
	if err != nil {
		return err
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

		case stakewise.InvalidVaultKey:
			// Invalid vault
			return stakewise.ErrInvalidVault
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return common.ErrInvalidPermissions
		}
	}
	return fmt.Errorf("nodeset server responded to validators-patch request with code %d: [%s]", code, response.Message)
}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node
func V2SubmitValidators_Get(c *common.CommonNodeSetClient, ctx context.Context, logger *slog.Logger, params map[string]string, validatorsPath string) (int, *common.NodeSetResponse[ValidatorsData], error) {
	// Send the request
	code, response, err := common.SubmitRequest[ValidatorsData](c, ctx, logger, true, http.MethodGet, nil, params, validatorsPath)
	if err != nil {
		return code, nil, fmt.Errorf("error getting registered validators: %w", err)
	}

	// Handle common errors
	switch code {
	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return code, nil, common.ErrInvalidSession
		}
	}
	return code, &response, nil
}
