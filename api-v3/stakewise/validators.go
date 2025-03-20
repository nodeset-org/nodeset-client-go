package v3stakewise

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
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

func (c *V3StakeWiseClient) VaultsValidator_Post(
	ctx context.Context,
	logger *slog.Logger,
	deployment string,
	vault ethcommon.Address,
	validatorRegistryRoot string,
	deadline uint64,
	validators string,
	signature string,
	exitSignatureIpsfHash string,
) (PostVaultsValidatorData, error) {
	// Create the request body
	request := VaultsValidatorPostRequest{
		// ValidatorsRegistryRoot: validatorRegistryRoot,
		// Deadline:               deadline,
		// Validators:             validators,
		// Signature:              signature,
		// ExitSignatureIpsfHash:  exitSignatureIpsfHash,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return PostVaultsValidatorData{}, fmt.Errorf("error marshalling vaults validator post request: %w", err)
	}
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + ValidatorsPath
	code, response, err := common.SubmitRequest[PostVaultsValidatorData](c.commonClient, ctx, logger, true, http.MethodPost, bytes.NewBuffer(jsonData), nil, path)
	if err != nil {
		return PostVaultsValidatorData{}, fmt.Errorf("error requesting minipool deposit signature: %w", err)
	}
	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Successfully generated minipool deposit signature
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return PostVaultsValidatorData{}, common.ErrInvalidDeployment
		case common.InvalidVaultKey:
			// Invalid vault
			return PostVaultsValidatorData{}, common.ErrInvalidVault
		}
	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid session
			return PostVaultsValidatorData{}, common.ErrInvalidSession
		}
	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return PostVaultsValidatorData{}, common.ErrInvalidPermissions
		}
	case http.StatusUnprocessableEntity:
		switch response.Error {
		case common.InsufficientVaultBalanceKey:
			// The vault doesn't have enough ETH deposits in it to support the number of validators being registered.
			return PostVaultsValidatorData{}, common.ErrInsufficientVaultBalance
		}
	}

	return PostVaultsValidatorData{}, fmt.Errorf("nodeset server responded to vaults validator request with code %d: [%s]", code, response.Message)

}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node on the provided deployment and vault
func (c *V3StakeWiseClient) Validators_Get(ctx context.Context, logger *slog.Logger, deployment string, vault ethcommon.Address) (stakewise.ValidatorsData, error) {
	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.ValidatorsPath
	code, response, err := stakewise.Validators_Get(c.commonClient, ctx, logger, nil, path)
	if err != nil {
		return stakewise.ValidatorsData{}, err
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return stakewise.ValidatorsData{}, common.ErrInvalidDeployment

		case stakewise.InvalidVaultKey:
			// Invalid vault
			return stakewise.ValidatorsData{}, stakewise.ErrInvalidVault
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return stakewise.ValidatorsData{}, common.ErrInvalidSession
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return stakewise.ValidatorsData{}, common.ErrInvalidPermissions
		}
	}
	return stakewise.ValidatorsData{}, fmt.Errorf("nodeset server responded to validators-get request with code %d: [%s]", code, response.Message)
}

// Submit signed exit data to NodeSet
func (c *V3StakeWiseClient) Validators_Patch(ctx context.Context, logger *slog.Logger, deployment string, vault ethcommon.Address, exitData []common.EncryptedExitData) error {
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
		case common.MalformedInputKey:
			// Invalid input
			return common.ErrMalformedInput
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return common.ErrInvalidDeployment
		case stakewise.InvalidVaultKey:
			// Invalid vault
			return stakewise.ErrInvalidVault
		case common.InvalidValidatorOwnerKey:
			// Invalid validator owner
			return common.ErrInvalidValidatorOwner
		case common.InvalidExitMessageKey:
			// Invalid exit message
			return common.ErrInvalidExitMessage
		}
	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return common.ErrInvalidSession
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
