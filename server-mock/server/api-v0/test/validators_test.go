package v0server_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
	"github.com/nodeset-org/nodeset-client-go/common"
	clientcommon "github.com/nodeset-org/nodeset-client-go/common"
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
	db := mgr.GetDatabase()
	deployment := db.StakeWise.AddDeployment(test.Network, test.ChainIDBig)
	_ = deployment.AddVault(test.StakeWiseVaultName, test.StakeWiseVaultAddress)
	node0Key, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	user, err := db.Core.AddUser(test.User0Email)
	require.NoError(t, err)
	node := user.WhitelistNode(node0Pubkey)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key, apiv0.NodeAddressMessageFormat)
	require.NoError(t, err)
	err = node.Register(regSig, apiv0.NodeAddressMessageFormat)
	require.NoError(t, err)

	// Create a session
	session := db.Core.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node0Pubkey, node0Key)
	require.NoError(t, err)
	err = db.Core.Login(node0Pubkey, session.Nonce, loginSig)
	require.NoError(t, err)

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
	session := db.Core.GetSessions()[0]

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
	pubkeys := make([]beacon.ValidatorPubkey, len(depositData))
	for i, data := range depositData {
		pubkeys[i] = beacon.ValidatorPubkey(data.PublicKey)
	}
	expectedData := map[beacon.ValidatorPubkey]apiv0.ValidatorStatus{
		pubkeys[0]: {
			Pubkey:              pubkeys[0],
			Status:              apiv0.StakeWiseStatus_Pending,
			ExitMessageUploaded: false,
		},
		pubkeys[1]: {
			Pubkey:              pubkeys[1],
			Status:              apiv0.StakeWiseStatus_Pending,
			ExitMessageUploaded: false,
		},
	}
	validatorsData := runGetValidatorsRequest(t, session)
	validatorMap := map[beacon.ValidatorPubkey]apiv0.ValidatorStatus{}
	for _, validator := range validatorsData.Validators {
		validatorMap[validator.Pubkey] = validator
	}
	require.Equal(t, expectedData, validatorMap)
	t.Logf("Received matching response")

	// Generate a signed exit for validator 1
	signedExit1 := idb.GenerateSignedExit(t, 1)
	t.Log("Generated signed exit")

	// Upload it
	runUploadSignedExitsRequest(t, session, []common.ExitData{signedExit1})
	t.Logf("Uploaded signed exit")

	// Get the validator status again
	validator := expectedData[pubkeys[1]]
	validator.ExitMessageUploaded = true
	expectedData[pubkeys[1]] = validator
	validatorsData = runGetValidatorsRequest(t, session)
	validatorMap = map[beacon.ValidatorPubkey]apiv0.ValidatorStatus{}
	for _, validator := range validatorsData.Validators {
		validatorMap[validator.Pubkey] = validator
	}
	require.Equal(t, expectedData, validatorMap)
	t.Logf("Received matching response")
}

// Run a GET api/validators request
func runGetValidatorsRequest(t *testing.T, session *db.Session) apiv0.ValidatorsData {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.Validators_Get(context.Background(), logger, test.Network)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a PATCH api/validators request
func runUploadSignedExitsRequest(t *testing.T, session *db.Session, signedExits []clientcommon.ExitData) {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	err := client.Validators_Patch(context.Background(), logger, signedExits, test.Network)
	require.NoError(t, err)
	t.Logf("Ran request")
}
