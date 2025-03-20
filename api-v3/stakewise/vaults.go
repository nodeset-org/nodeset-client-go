package v3stakewise

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	stakewise "github.com/nodeset-org/nodeset-client-go/common/stakewise"
)

const (
	VaultsPath     string = "vaults"
	ValidatorsPath string = "validators"
	MetaPath       string = "meta"
)

type VaultsData struct {
	Vaults []ethcommon.Address `json:"vaults"`
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

// Returns information about the requesting user's node account with respect to the number of validators the user has deployed and can deploy on this vault.
func (c *V3StakeWiseClient) ValidatorMeta_Get(ctx context.Context, logger *slog.Logger, deployment string, vault ethcommon.Address) (stakewise.VaultsMetaData, error) {
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + ValidatorsPath + "/" + MetaPath
	code, response, err := common.SubmitRequest[stakewise.VaultsMetaData](c.commonClient, ctx, logger, true, http.MethodGet, nil, nil, path)
	if err != nil {
		return stakewise.VaultsMetaData{}, fmt.Errorf("error submitting vaults request: %w", err)
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
			return stakewise.VaultsMetaData{}, common.ErrInvalidDeployment
		case common.InvalidVaultKey:
			// Invalid vault
			return stakewise.VaultsMetaData{}, common.ErrInvalidVault
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return stakewise.VaultsMetaData{}, common.ErrInvalidSession
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return stakewise.VaultsMetaData{}, common.ErrInvalidPermissions
		}
	}

	return stakewise.VaultsMetaData{}, fmt.Errorf("nodeset server responded to vaults validator meta request with code %d: [%s]", code, response.Message)
}
