package v2stakewise

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common"
)

// Gets the list of deployments available on the server
func (c *V2StakeWiseClient) Deployments(ctx context.Context) (common.DeploymentsData, error) {
	// Submit the request
	code, response, err := common.SubmitRequest[common.DeploymentsData](c.commonClient, ctx, true, http.MethodGet, nil, nil, StakeWisePrefix+common.DeploymentsPath)
	if err != nil {
		return common.DeploymentsData{}, fmt.Errorf("error submitting deployments request: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Success
		return response.Data, nil

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return common.DeploymentsData{}, common.ErrInvalidSession
		}
	}
	return common.DeploymentsData{}, fmt.Errorf("nodeset server responded to deployments request with code %d: [%s]", code, response.Message)
}
