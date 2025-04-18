package v2stakewise

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"path"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
)

const (
	VaultsPath string = "vaults"
)

type VaultsData struct {
	Vaults []ethcommon.Address `json:"vaults"`
}

// Gets the list of vaults available on the server for the provided deployment
func (c *V2StakeWiseClient) Vaults(ctx context.Context, logger *slog.Logger, deployment string) (VaultsData, error) {
	// Submit the request
	pathString := path.Join(StakeWisePrefix, deployment, VaultsPath)
	code, response, err := common.SubmitRequest[VaultsData](c.commonClient, ctx, logger, true, http.MethodGet, nil, nil, pathString)
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
