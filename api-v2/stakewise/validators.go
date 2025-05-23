package v2stakewise

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/rocket-pool/node-manager-core/beacon"
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
	Pubkey              beacon.ValidatorPubkey `json:"pubkey"`
	Status              StakeWiseStatus        `json:"status"`
	ExitMessageUploaded bool                   `json:"exitMessage"`
}

// Response to a validators request
type ValidatorsData struct {
	Validators []ValidatorStatus `json:"validators"`
}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node on the provided deployment and vault
func (c *V2StakeWiseClient) Validators_Get(ctx context.Context, logger *slog.Logger, deployment string, vault ethcommon.Address) (ValidatorsData, error) {
	// Send the request
	pathString := path.Join(StakeWisePrefix, deployment, vault.Hex(), stakewise.ValidatorsPath)
	code, response, err := stakewise.Validators_Get[ValidatorsData](c.commonClient, ctx, logger, nil, pathString)
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

		case common.InvalidVaultKey:
			// Invalid vault
			return ValidatorsData{}, common.ErrInvalidVault
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
	pathString := path.Join(StakeWisePrefix, deployment, vault.Hex(), stakewise.ValidatorsPath)
	code, response, err := stakewise.Validators_Patch(c.commonClient, ctx, logger, body, nil, pathString)
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

		case common.InvalidVaultKey:
			// Invalid vault
			return common.ErrInvalidVault
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
