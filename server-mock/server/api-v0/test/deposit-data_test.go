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
	vault := db.StakeWise.GetDeployment(test.Network).GetStakeWiseVault(test.StakeWiseVaultAddress)
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
	expectedData := []stakewise.ValidatorStatus{
		{
			Pubkey:              beacon.ValidatorPubkey(depositData[0].PublicKey),
			Status:              stakewise.StakeWiseStatus_Pending,
			ExitMessageUploaded: false,
		},
		{
			Pubkey:              beacon.ValidatorPubkey(depositData[1].PublicKey),
			Status:              stakewise.StakeWiseStatus_Pending,
			ExitMessageUploaded: false,
		},
		{
			Pubkey:              beacon.ValidatorPubkey(depositData[2].PublicKey),
			Status:              stakewise.StakeWiseStatus_Pending,
			ExitMessageUploaded: false,
		},
	}
	require.Equal(t, expectedData, validatorsData.Validators)
	t.Logf("Received matching response")
}

// Run a GET api/deposit-data request
func runGetDepositDataRequest(t *testing.T, session *db.Session) stakewise.DepositDataData {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.DepositData_Get(context.Background(), test.StakeWiseVaultAddress, test.Network)
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
	err := client.DepositData_Post(context.Background(), depositData)
	require.NoError(t, err)
	t.Logf("Ran request")
}
