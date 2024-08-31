package apiv0

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/utils"
)

const (
	// Deposit data has withdrawal creds that don't match a StakeWise vault
	VaultNotFoundKey string = "vault_not_found"
)

var (
	// The requested StakeWise vault didn't exist
	ErrVaultNotFound error = errors.New("deposit data has withdrawal creds that don't match a StakeWise vault")
)

// Get the current version of the aggregated deposit data on the server
func (c *NodeSetClient) DepositDataMeta(ctx context.Context, vault ethcommon.Address, network string) (stakewise.DepositDataMetaData, error) {
	// Create the request params
	vaultString := utils.RemovePrefix(strings.ToLower(vault.Hex()))
	params := map[string]string{
		"vault":   vaultString,
		"network": network,
	}

	// Send it
	code, response, err := stakewise.DepositDataMeta(c.CommonNodeSetClient, ctx, params, stakewise.DepositDataMetaPath)
	if err != nil {
		return stakewise.DepositDataMetaData{}, err
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case InvalidNetworkKey:
			// Network not known
			return stakewise.DepositDataMetaData{}, ErrInvalidNetwork
		}
	}
	return stakewise.DepositDataMetaData{}, fmt.Errorf("nodeset server responded to deposit-data-meta request with code %d: [%s]", code, response.Message)
}

// Get the aggregated deposit data from the server
func (c *NodeSetClient) DepositData_Get(ctx context.Context, vault ethcommon.Address, network string) (stakewise.DepositDataData, error) {
	// Create the request params
	vaultString := utils.RemovePrefix(strings.ToLower(vault.Hex()))
	params := map[string]string{
		"vault":   vaultString,
		"network": network,
	}

	// Send it
	code, response, err := stakewise.DepositData_Get[stakewise.DepositDataData](c.CommonNodeSetClient, ctx, params, stakewise.DepositDataPath)
	if err != nil {
		return stakewise.DepositDataData{}, err
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case InvalidNetworkKey:
			// Network not known
			return stakewise.DepositDataData{}, ErrInvalidNetwork
		}
	}
	return stakewise.DepositDataData{}, fmt.Errorf("nodeset server responded to deposit-data-get request with code %d: [%s]", code, response.Message)
}

// Uploads deposit data to NodeSet
func (c *NodeSetClient) DepositData_Post(ctx context.Context, depositData []beacon.ExtendedDepositData) error {
	// Send the request
	code, response, err := stakewise.DepositData_Post(c.CommonNodeSetClient, ctx, depositData, stakewise.DepositDataPath)
	if err != nil {
		return err
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return nil

	case http.StatusBadRequest:
		switch response.Error {
		case VaultNotFoundKey:
			// The requested StakeWise vault didn't exist
			return ErrVaultNotFound
		}
	}
	return fmt.Errorf("nodeset server responded to deposit-data-post request with code %d: [%s]", code, response.Message)
}

// Get a list of all of the pubkeys that have already been registered with NodeSet for this node
func (c *NodeSetClient) Validators_Get(ctx context.Context, network string) (stakewise.ValidatorsData, error) {
	// Create the request params
	queryParams := map[string]string{
		"network": network,
	}

	// Send the request
	code, response, err := stakewise.Validators_Get(c.CommonNodeSetClient, ctx, queryParams, stakewise.ValidatorsPath)
	if err != nil {
		return stakewise.ValidatorsData{}, err
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case InvalidNetworkKey:
			// Network not known
			return stakewise.ValidatorsData{}, ErrInvalidNetwork
		}
	}
	return stakewise.ValidatorsData{}, fmt.Errorf("nodeset server responded to validators-get request with code %d: [%s]", code, response.Message)
}

// Submit signed exit data to NodeSet
func (c *NodeSetClient) Validators_Patch(ctx context.Context, exitData []common.ExitData, network string) error {
	// Create the request params
	params := map[string]string{
		"network": network,
	}

	// Submit the request
	code, response, err := stakewise.Validators_Patch(c.CommonNodeSetClient, ctx, exitData, params, stakewise.ValidatorsPath)
	if err != nil {
		return err
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return nil

	case http.StatusBadRequest:
		switch response.Error {
		case InvalidNetworkKey:
			// Network not known
			return ErrInvalidNetwork
		}
	}
	return fmt.Errorf("nodeset server responded to validators-patch request with code %d: [%s]", code, response.Message)
}
