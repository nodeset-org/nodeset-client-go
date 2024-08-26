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

// Extended deposit data beyond what is required in an actual deposit message to Beacon, emulating what the deposit CLI produces
type ExtendedDepositData struct {
	PublicKey             beacon.ByteArray `json:"pubkey"`
	WithdrawalCredentials beacon.ByteArray `json:"withdrawalCredentials"`
	Amount                uint64           `json:"amount"`
	Signature             beacon.ByteArray `json:"signature"`
	DepositMessageRoot    beacon.ByteArray `json:"depositMessageRoot"`
	DepositDataRoot       beacon.ByteArray `json:"depositDataRoot"`
	ForkVersion           beacon.ByteArray `json:"forkVersion"`
	NetworkName           string           `json:"networkName"`
}

// Response to a deposit data request
type DepositDataData struct {
	Version     int                   `json:"version"`
	DepositData []ExtendedDepositData `json:"depositData"`
}

// Request body for uploading deposit data
type DepositData_PostBody struct {
	Validators []ExtendedDepositData `json:"validators"`
}

// Get the aggregated deposit data from the server
func (c *V2StakeWiseClient) DepositData_Get(ctx context.Context, deployment string, vault ethcommon.Address) (stakewise.DepositDataData, error) {
	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.DepositDataPath
	code, response, err := stakewise.DepositData_Get[DepositDataData](c.commonClient, ctx, nil, path)
	if err != nil {
		return stakewise.DepositDataData{}, err
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Convert to standard Beacon data
		data := stakewise.DepositDataData{
			Version:     response.Data.Version,
			DepositData: make([]beacon.ExtendedDepositData, len(response.Data.DepositData)),
		}
		for i, deposit := range response.Data.DepositData {
			data.DepositData[i] = beacon.ExtendedDepositData(deposit)
		}
		return data, nil

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
	// Convert the deposit data to the NS form
	body := DepositData_PostBody{}
	body.Validators = make([]ExtendedDepositData, len(depositData))
	for i, deposit := range depositData {
		body.Validators[i] = ExtendedDepositData(deposit)
	}

	// Send the request
	path := StakeWisePrefix + deployment + "/" + vault.Hex() + "/" + stakewise.DepositDataPath
	code, response, err := stakewise.DepositData_Post(c.commonClient, ctx, body, path)
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
