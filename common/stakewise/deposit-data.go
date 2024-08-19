package stakewise

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

const (
	// Route for getting the latest deposit data set from the NodeSet server
	DepositDataPath string = "deposit-data"

	// Subroute for getting the version of the latest deposit data
	DepositDataMetaPath string = DepositDataPath + "/meta"

	// Deposit data can't be uploaded to Mainnet because the user isn't allowed to use Mainnet yet
	InvalidPermissionsKey string = "invalid_permissions"
)

var (
	// The user isn't allowed to use the provided vault yet
	ErrInvalidPermissions error = errors.New("deposit data can't be uploaded to the specified vault because you aren't permitted to use it yet")
)

// Response to a deposit data request
type DepositDataData struct {
	Version     int                          `json:"version"`
	DepositData []beacon.ExtendedDepositData `json:"depositData"`
}

// Response to a deposit data meta request
type DepositDataMetaData struct {
	Version int `json:"version"`
}

// Get the aggregated deposit data from the server
func DepositData_Get(c *common.CommonNodeSetClient, ctx context.Context, params map[string]string, depositDataPath string) (int, *common.NodeSetResponse[DepositDataData], error) {
	// Send the request
	code, response, err := common.SubmitRequest[DepositDataData](c, ctx, true, http.MethodGet, nil, params, depositDataPath)
	if err != nil {
		return code, nil, fmt.Errorf("error getting deposit data: %w", err)
	}

	// Handle common errors
	switch code {
	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return code, nil, common.ErrInvalidSession
		}
	}

	return code, &response, nil
}

// Get the current version of the aggregated deposit data on the server
func DepositDataMeta(c *common.CommonNodeSetClient, ctx context.Context, params map[string]string, depositDataMetaPath string) (int, *common.NodeSetResponse[DepositDataMetaData], error) {
	// Send the request
	code, response, err := common.SubmitRequest[DepositDataMetaData](c, ctx, true, http.MethodGet, nil, params, depositDataMetaPath)
	if err != nil {
		return code, nil, fmt.Errorf("error getting deposit data version: %w", err)
	}

	// Handle common errors
	switch code {
	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return code, nil, common.ErrInvalidSession
		}
	}
	return code, &response, nil
}

// Uploads deposit data to Nodeset
func DepositData_Post(c *common.CommonNodeSetClient, ctx context.Context, depositData []beacon.ExtendedDepositData, depositDataPath string) (int, *common.NodeSetResponse[struct{}], error) {
	// Create the request body
	serializedData, err := json.Marshal(depositData)
	if err != nil {
		return -1, nil, fmt.Errorf("error serializing deposit data: %w", err)
	}

	// Send it
	code, response, err := common.SubmitRequest[struct{}](c, ctx, true, http.MethodPost, bytes.NewBuffer(serializedData), nil, depositDataPath)
	if err != nil {
		return code, nil, fmt.Errorf("error uploading deposit data: %w", err)
	}

	// Handle common errors
	switch code {
	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return code, nil, common.ErrInvalidSession
		}

	case http.StatusForbidden:
		switch response.Error {
		case InvalidPermissionsKey:
			// The user isn't allowed to use the vault yet
			return code, nil, ErrInvalidPermissions
		}
	}
	return code, &response, nil
}
