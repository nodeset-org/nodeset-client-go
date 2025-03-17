package v3stakewise

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
)

// Get the current version of the aggregated deposit data on the server
func (c *V3StakeWiseClient) DepositDataMeta(ctx context.Context, logger *slog.Logger, deployment string, vault ethcommon.Address) (stakewise.DepositDataMetaData, error) {
	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.DepositDataMetaPath
	code, response, err := stakewise.DepositDataMeta(c.commonClient, ctx, logger, nil, path)
	if err != nil {
		return stakewise.DepositDataMetaData{}, err
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return stakewise.DepositDataMetaData{}, common.ErrInvalidDeployment

		case stakewise.InvalidVaultKey:
			// Invalid vault
			return stakewise.DepositDataMetaData{}, stakewise.ErrInvalidVault
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return stakewise.DepositDataMetaData{}, common.ErrInvalidPermissions
		}
	}
	return stakewise.DepositDataMetaData{}, fmt.Errorf("nodeset server responded to deposit-data-meta request with code %d: [%s]", code, response.Message)
}
