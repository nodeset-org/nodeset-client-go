package v0server_test

import (
	"context"
	"fmt"
	"testing"

	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	idb "github.com/nodeset-org/nodeset-client-go/server-mock/internal/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/stretchr/testify/require"
)

// Make sure the correct response is returned for a successful request
func TestGetDepositData(t *testing.T) {
	// Take a snapshot
	mgr.TakeSnapshot("test")
	defer func() {
		err := mgr.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	db := idb.ProvisionFullDatabase(t, logger, true)
	mgr.SetDatabase(db)

	// Run a get deposit data request
	data := runGetDepositDataRequest(t, db.Core.GetSessions()[0])

	// Make sure the response is correct
	vault := db.StakeWise.GetDeployment(test.Network).GetVault(test.StakeWiseVaultAddress)
	require.Equal(t, vault.LatestDepositDataSetIndex, data.Version)
	require.Equal(t, vault.LatestDepositDataSet, data.DepositData)
	require.Greater(t, len(data.DepositData), 0)
	t.Logf("Received correct response - version = %d, deposit data matches", data.Version)
}

// Make sure the deposit data is uploaded correctly
func TestUploadDepositData(t *testing.T) {
	// Take a snapshot
	mgr.TakeSnapshot("test")
	defer func() {
		err := mgr.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	db := idb.ProvisionFullDatabase(t, logger, false)
	mgr.SetDatabase(db)
	session := db.Core.GetSessions()[0]

	// Run a get deposit data request to make sure it's empty
	data := runGetDepositDataRequest(t, session)
	require.Equal(t, 0, data.Version)
	require.Empty(t, data.DepositData)

	// Generate new deposit data
	depositData := []beacon.ExtendedDepositData{
		idb.GenerateDepositData(t, 0, test.StakeWiseVaultAddress),
		idb.GenerateDepositData(t, 1, test.StakeWiseVaultAddress),
		idb.GenerateDepositData(t, 2, test.StakeWiseVaultAddress),
	}
	t.Log("Generated deposit data")

	// Run an upload deposit data request
	runUploadDepositDataRequest(t, session, depositData)

	// Run a get deposit data request to make sure it's uploaded
	validatorsData := runGetValidatorsRequest(t, db.Core.GetSessions()[0])
	validatorMap := map[beacon.ValidatorPubkey]stakewise.ValidatorStatus{}
	for _, validator := range validatorsData.Validators {
		validatorMap[validator.Pubkey] = validator
	}

	pubkey0 := beacon.ValidatorPubkey(depositData[0].PublicKey)
	pubkey1 := beacon.ValidatorPubkey(depositData[1].PublicKey)
	pubkey2 := beacon.ValidatorPubkey(depositData[2].PublicKey)
	expectedMap := map[beacon.ValidatorPubkey]stakewise.ValidatorStatus{
		pubkey0: {
			Pubkey:              pubkey0,
			Status:              stakewise.StakeWiseStatus_Pending,
			ExitMessageUploaded: false,
		},
		pubkey1: {
			Pubkey:              pubkey1,
			Status:              stakewise.StakeWiseStatus_Pending,
			ExitMessageUploaded: false,
		},
		pubkey2: {
			Pubkey:              pubkey2,
			Status:              stakewise.StakeWiseStatus_Pending,
			ExitMessageUploaded: false,
		},
	}
	require.Equal(t, expectedMap, validatorMap)
	t.Logf("Received matching response")
}

// Run a GET api/deposit-data request
func runGetDepositDataRequest(t *testing.T, session *db.Session) stakewise.DepositDataData {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.DepositData_Get(context.Background(), logger, test.StakeWiseVaultAddress, test.Network)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a POST api/deposit-data request
func runUploadDepositDataRequest(t *testing.T, session *db.Session, depositData []beacon.ExtendedDepositData) {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	err := client.DepositData_Post(context.Background(), logger, depositData)
	require.NoError(t, err)
	t.Logf("Ran request")
}
