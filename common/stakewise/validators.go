package stakewise

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/goccy/go-json"
	"github.com/nodeset-org/nodeset-client-go/common"
)

const (
	// Route for interacting with the list of validators
	ValidatorsPath string = "validators"

	ValidatorsMetaPath string = ValidatorsPath + "/meta"
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

// Submit signed exit data to Nodeset
func Validators_Patch(c *common.CommonNodeSetClient, ctx context.Context, logger *slog.Logger, exitData any, params map[string]string, validatorsPath string) (int, *common.NodeSetResponse[struct{}], error) {
	// Create the request body
	jsonData, err := json.Marshal(exitData)
	if err != nil {
		return -1, nil, fmt.Errorf("error marshalling exit data to JSON: %w", err)
	}

	// Submit the request
	code, response, err := common.SubmitRequest[struct{}](c, ctx, logger, true, http.MethodPatch, bytes.NewBuffer(jsonData), params, validatorsPath)
	if err != nil {
		return code, nil, fmt.Errorf("error submitting exit data: %w", err)
	}

	// Handle common errors
	switch code {
	case http.StatusBadRequest:
		switch response.Error {
		case common.MalformedInputKey:
			// Invalid input
			return code, nil, common.ErrMalformedInput

		case common.InvalidValidatorOwnerKey:
			// Invalid validator owner
			return code, nil, common.ErrInvalidValidatorOwner

		case common.InvalidExitMessageKey:
			// Invalid exit message
			return code, nil, common.ErrInvalidExitMessage
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return code, nil, common.ErrInvalidSession
		}
	}
	return code, &response, nil
}

func Validators_Get[T any](
	c *common.CommonNodeSetClient,
	ctx context.Context,
	logger *slog.Logger,
	params map[string]string,
	validatorsPath string,
) (int, *common.NodeSetResponse[T], error) {
	// Send the request
	code, response, err := common.SubmitRequest[T](c, ctx, logger, true, http.MethodGet, nil, params, validatorsPath)
	if err != nil {
		return code, nil, fmt.Errorf("error getting registered validators: %w", err)
	}

	// Handle common errors
	switch code {
	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			return code, nil, common.ErrInvalidSession
		}
	}

	return code, &response, nil
}
