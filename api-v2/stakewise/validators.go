package v2stakewise

import (
	"context"
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
)

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node on the provided deployment and vault
func (c *V2StakeWiseClient) Validators_Get(ctx context.Context, deployment string, vault ethcommon.Address) (stakewise.ValidatorsData, error) {
	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.ValidatorsPath
	code, response, err := stakewise.Validators_Get(c.commonClient, ctx, nil, path)
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
	}
	return stakewise.ValidatorsData{}, fmt.Errorf("nodeset server responded to validators-get request with code %d: [%s]", code, response.Message)
}

// Submit signed exit data to NodeSet
func (c *V2StakeWiseClient) Validators_Patch(ctx context.Context, deployment string, vault ethcommon.Address, exitData []common.ExitData) error {
	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.ValidatorsPath
	code, response, err := stakewise.Validators_Patch(c.commonClient, ctx, exitData, nil, path)
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
	}
	return fmt.Errorf("nodeset server responded to validators-patch request with code %d: [%s]", code, response.Message)
}
