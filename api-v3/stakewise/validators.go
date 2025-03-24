package v3stakewise

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/goccy/go-json"

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

type Validators_PostBody struct {
	Validators        []ValidatorRegistrationDetails `json:"validators"`
	BeaconDepositRoot ethcommon.Hash                 `json:"beaconDepositRoot"`
}

type PostValidatorData struct {
	Signature string `json:"signature"`
}

type ValidatorRegistrationDetails struct {
	DepositData beacon.ExtendedDepositData `json:"depositData"`
	ExitMessage string                     `json:"exitMessage"`
}

// Validator status info
type ValidatorStatus struct {
	Pubkey              beacon.ValidatorPubkey `json:"pubkey"`
	ExitMessageUploaded bool                   `json:"exitMessage"`
}

// Response to a validators request
type ValidatorsData struct {
	Validators []ValidatorStatus `json:"validators"`
}

func (c *V3StakeWiseClient) Validators_Post(
	ctx context.Context,
	logger *slog.Logger,
	deployment string,
	vault ethcommon.Address,
	validators []ValidatorRegistrationDetails,
	beaconDepositRoot ethcommon.Hash,
) (PostValidatorData, error) {
	// Create the request body
	request := Validators_PostBody{
		Validators:        validators,
		BeaconDepositRoot: beaconDepositRoot,
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return PostValidatorData{}, fmt.Errorf("error marshalling validator post request: %w", err)
	}
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + ValidatorsPath
	code, response, err := common.SubmitRequest[PostValidatorData](c.commonClient, ctx, logger, true, http.MethodPost, bytes.NewBuffer(jsonData), nil, path)
	if err != nil {
		return PostValidatorData{}, fmt.Errorf("error requesting validator manager signature: %w", err)
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
			return PostValidatorData{}, common.ErrInvalidDeployment
		case common.InvalidVaultKey:
			// Invalid vault
			return PostValidatorData{}, common.ErrInvalidVault
		}
	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid session
			return PostValidatorData{}, common.ErrInvalidSession
		}
	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return PostValidatorData{}, common.ErrInvalidPermissions
		}
	case http.StatusUnprocessableEntity:
		switch response.Error {
		case common.InsufficientVaultBalanceKey:
			// The vault doesn't have enough ETH deposits in it to support the number of validators being registered.
			return PostValidatorData{}, common.ErrInsufficientVaultBalance
		}
	}

	return PostValidatorData{}, fmt.Errorf("nodeset server responded to vaults validator request with code %d: [%s]", code, response.Message)

}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node on the provided deployment and vault
func (c *V3StakeWiseClient) Validators_Get(ctx context.Context, logger *slog.Logger, deployment string, vault ethcommon.Address) (ValidatorsData, error) {
	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.ValidatorsPath
	code, response, err := V3SubmitValidators_Get(c.commonClient, ctx, logger, nil, path)
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

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node
func V3SubmitValidators_Get(c *common.CommonNodeSetClient, ctx context.Context, logger *slog.Logger, params map[string]string, validatorsPath string) (int, *common.NodeSetResponse[ValidatorsData], error) {
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
