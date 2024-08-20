package v2stakewise

import (
	"context"
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	// The provided deposit data does not match the given deployment or vault
	DepositDataMismatchKey string = "deposit_data_mismatch"
)

var (
	// The provided deposit data does not match the given deployment or vault
	ErrDepositDataMismatch error = fmt.Errorf("the provided deposit data does not match the given deployment or vault")
)

// Get the aggregated deposit data from the server
func (c *V2StakeWiseClient) DepositData_Get(ctx context.Context, deployment string, vault ethcommon.Address) (stakewise.DepositDataData, error) {
	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.DepositDataPath
	code, response, err := stakewise.DepositData_Get(c.commonClient, ctx, nil, path)
	if err != nil {
		return stakewise.DepositDataData{}, err
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return stakewise.DepositDataData{}, common.ErrInvalidDeployment

		case stakewise.InvalidVaultKey:
			// Invalid vault
			return stakewise.DepositDataData{}, stakewise.ErrInvalidVault
		}
	}
	return stakewise.DepositDataData{}, fmt.Errorf("nodeset server responded to deposit-data-get request with code %d: [%s]", code, response.Message)
}

// Uploads deposit data to NodeSet
func (c *V2StakeWiseClient) DepositData_Post(ctx context.Context, deployment string, vault ethcommon.Address, depositData []beacon.ExtendedDepositData) error {
	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.DepositDataPath
	code, response, err := stakewise.DepositData_Post(c.commonClient, ctx, depositData, path)
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

		case common.MalformedInputKey:
			// Malformed input
			return common.ErrMalformedInput

		case DepositDataMismatchKey:
			// Deposit data mismatch
			return ErrDepositDataMismatch
		}
	}
	return fmt.Errorf("nodeset server responded to deposit-data-post request with code %d: [%s]", code, response.Message)
}
