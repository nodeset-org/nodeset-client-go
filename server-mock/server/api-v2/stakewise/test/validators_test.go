package v2server_stakewise_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/common"
	clientcommon "github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	idb "github.com/nodeset-org/nodeset-client-go/server-mock/internal/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/stretchr/testify/require"
)

// Make sure the correct response is returned for a successful request
func TestGetValidators(t *testing.T) {
	// Take a snapshot
	mgr.TakeSnapshot("test")
	defer func() {
		err := mgr.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	node0Key, err := test.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	err = mgr.AddUser(test.User0Email)
	if err != nil {
		t.Fatalf("error adding user: %v", err)
	}
	err = mgr.WhitelistNodeAccount(test.User0Email, node0Pubkey)
	if err != nil {
		t.Fatalf("error whitelisting node account: %v", err)
	}
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key)
	if err != nil {
		t.Fatalf("error getting signature for registration: %v", err)
	}
	err = mgr.RegisterNodeAccount(test.User0Email, node0Pubkey, regSig)
	if err != nil {
		t.Fatalf("error registering node account: %v", err)
	}

	// Create a session
	session := mgr.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node0Pubkey, node0Key)
	if err != nil {
		t.Fatalf("error getting signature for login: %v", err)
	}
	err = mgr.Login(session.Nonce, node0Pubkey, loginSig)
	if err != nil {
		t.Fatalf("error logging in: %v", err)
	}

	// Run a get validators request
	data := runGetValidatorsRequest(t, session)

	// Make sure the response is correct
	require.Empty(t, data.Validators)
	t.Logf("Received correct response - validators is empty")
}

// Make sure signed exits are uploaded correctly
func TestUploadSignedExits(t *testing.T) {
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
	session := db.Sessions[0]

	// Run a get deposit data request to make sure it's empty
	data := runGetDepositDataRequest(t, session)
	require.Equal(t, 0, data.Version)
	require.Empty(t, data.DepositData)

	// Generate new deposit data
	depositData := []beacon.ExtendedDepositData{
		idb.GenerateDepositData(t, 0, test.StakeWiseVaultAddress),
		idb.GenerateDepositData(t, 1, test.StakeWiseVaultAddress),
	}
	t.Log("Generated deposit data")

	// Run an upload deposit data request
	runUploadDepositDataRequest(t, session, depositData)

	// Run a get deposit data request to make sure it's uploaded
	validatorsData := runGetValidatorsRequest(t, session)
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
	}
	require.Equal(t, expectedData, validatorsData.Validators)
	t.Logf("Received matching response")

	// Generate a signed exit for validator 1
	signedExit1 := idb.GenerateSignedExit(t, 1)
	t.Log("Generated signed exit")

	// Upload it
	runUploadSignedExitsRequest(t, session, []common.ExitData{signedExit1})
	t.Logf("Uploaded signed exit")

	// Get the validator status again
	validatorsData = runGetValidatorsRequest(t, session)
	expectedData = []stakewise.ValidatorStatus{
		{
			Pubkey:              beacon.ValidatorPubkey(depositData[0].PublicKey),
			Status:              stakewise.StakeWiseStatus_Pending,
			ExitMessageUploaded: false,
		},
		{
			Pubkey:              beacon.ValidatorPubkey(depositData[1].PublicKey),
			Status:              stakewise.StakeWiseStatus_Pending,
			ExitMessageUploaded: true, // This should be true now
		},
	}
	require.Equal(t, expectedData, validatorsData.Validators)
	t.Logf("Received matching response")
}

// Run a GET api/validators request
func runGetValidatorsRequest(t *testing.T, session *db.Session) stakewise.ValidatorsData {
	// Create the client
	client := apiv2.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.StakeWise.Validators_Get(context.Background(), test.Network, test.StakeWiseVaultAddress)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a PATCH api/validators request
func runUploadSignedExitsRequest(t *testing.T, session *db.Session, signedExits []clientcommon.ExitData) {
	// Create the client
	client := apiv2.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	err := client.StakeWise.Validators_Patch(context.Background(), test.Network, test.StakeWiseVaultAddress, signedExits)
	require.NoError(t, err)
	t.Logf("Ran request")
}
