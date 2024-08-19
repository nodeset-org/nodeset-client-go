package apiv0

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
)

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

// Submit signed exit data to Nodeset
func (c *NodeSetClient) Validators_Patch(ctx context.Context, exitData []common.ExitData, network string) error {
	// Create the request params
	params := map[string]string{
		"network": network,
	}

	// Submit the request
	code, response, err := stakewise.Validators_Patch(c.CommonNodeSetClient, ctx, exitData, params, stakewise.ValidatorsPath)
	if err != nil {
		return fmt.Errorf("error submitting exit data: %w", err)
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
