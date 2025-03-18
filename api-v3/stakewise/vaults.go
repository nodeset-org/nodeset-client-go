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
)

const (
	VaultsPath     string = "vaults"
	ValidatorsPath string = "validators"
	MetaPath       string = "meta"
)

type VaultsData struct {
	Vaults []ethcommon.Address `json:"vaults"`
}

type VaultsMetaData struct {
	// validators that the user has for this vault that are active on the Beacon Chain (e.g., pending and active, *not* exited or slashed).
	Active uint `json:"active"`

	// validators that the current user is allowed to have for this vault
	Max uint `json:"max"`

	// validators the user is still permitted to create and upload to this vault.
	Available uint `json:"available"`
}

type PostVaultsValidatorData struct {
	Signature string `json:"signature"`
}

type DepositDataDetails struct {
	PublicKey             string `json:"pubkey"`
	WithdrawalCredentials string `json:"withdrawalCredentials"`
	Amount                uint   `json:"amount"`
	Signature             string `json:"signature"`
	DepositMessageRoot    string `json:"depositMessageRoot"`
	DepositDataRoot       string `json:"depositDataRoot"`
	ForkVersion           string `json:"forkVersion"`
	NetworkName           string `json:"networkName"`
}

type ValidatorRegistrationDetails struct {
	DepositData DepositDataDetails `json:"depositData"`
	ExitMessage string             `json:"exitMessage"`
}

type VaultsValidatorPostRequest struct {
	Validators []ValidatorRegistrationDetails `json:"validators"`
}

// Gets the list of vaults available on the server for the provided deployment
func (c *V3StakeWiseClient) Vaults(ctx context.Context, logger *slog.Logger, deployment string) (VaultsData, error) {
	// Submit the request
	path := StakeWisePrefix + deployment + "/" + VaultsPath
	code, response, err := common.SubmitRequest[VaultsData](c.commonClient, ctx, logger, true, http.MethodGet, nil, nil, path)
	if err != nil {
		return VaultsData{}, fmt.Errorf("error submitting vaults request: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Success
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return VaultsData{}, common.ErrInvalidDeployment
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return VaultsData{}, common.ErrInvalidSession
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return VaultsData{}, common.ErrInvalidPermissions
		}
	}
	return VaultsData{}, fmt.Errorf("nodeset server responded to vaults request with code %d: [%s]", code, response.Message)
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
		ValidatorsRegistryRoot: validatorRegistryRoot,
		Deadline:               deadline,
		Validators:             validators,
		Signature:              signature,
		ExitSignatureIpsfHash:  exitSignatureIpsfHash,
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

// Returns information about the requesting user's node account with respect to the number of validators the user has deployed and can deploy on this vault.
func (c *V3StakeWiseClient) VaultsValidatorMeta_Get(ctx context.Context, logger *slog.Logger, deployment string, vault ethcommon.Address) (VaultsMetaData, error) {
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + ValidatorsPath + "/" + MetaPath
	code, response, err := common.SubmitRequest[VaultsMetaData](c.commonClient, ctx, logger, true, http.MethodGet, nil, nil, path)
	if err != nil {
		return VaultsMetaData{}, fmt.Errorf("error submitting vaults request: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Success
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return VaultsMetaData{}, common.ErrInvalidDeployment
		case common.InvalidVaultKey:
			// Invalid vault
			return VaultsMetaData{}, common.ErrInvalidVault
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return VaultsMetaData{}, common.ErrInvalidSession
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return VaultsMetaData{}, common.ErrInvalidPermissions
		}
	}

	return VaultsMetaData{}, fmt.Errorf("nodeset server responded to vaults validator meta request with code %d: [%s]", code, response.Message)
}
