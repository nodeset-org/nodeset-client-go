package v3server_stakewise_test

import (
	"context"
	"fmt"
	"testing"

	apiv3 "github.com/nodeset-org/nodeset-client-go/api-v3"
	stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
	"github.com/rocket-pool/node-manager-core/beacon"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	v3core "github.com/nodeset-org/nodeset-client-go/api-v3/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

func TestPostValidators(t *testing.T) {
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
	deployment.AddVault(test.StakeWiseVaultAddress)
	deployment.MaxValidators = 10 // Set max validators

	node0Key, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	user, err := db.Core.AddUser(test.User0Email)
	require.NoError(t, err)
	node := user.WhitelistNode(node0Pubkey)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key, v3core.NodeAddressMessageFormat)
	require.NoError(t, err)
	err = node.Register(regSig, v3core.NodeAddressMessageFormat)
	require.NoError(t, err)

	// Create a session
	session := db.Core.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node0Pubkey, node0Key)
	require.NoError(t, err)
	err = db.Core.Login(node0Pubkey, session.Nonce, loginSig)
	require.NoError(t, err)

	// Get the initial validator limits
	metaBefore := runGetValidatorsMetaRequest(t, session)
	require.Equal(t, uint(0), metaBefore.Active)     // No validators yet
	require.Equal(t, uint(metaBefore.Max), uint(10)) // Max set to 10

	// Generate validator details
	numValidatorsToRegister := 3
	validatorDetails := make([]stakewise.ValidatorRegistrationDetails, numValidatorsToRegister)
	for i := 0; i < numValidatorsToRegister; i++ {
		validatorDetails[i] = stakewise.ValidatorRegistrationDetails{
			DepositData: beacon.ExtendedDepositData{},
			ExitMessage: fmt.Sprintf("exit_%d", i),
		}
	}

	// Submit the request (TODO)
	beaconDepositRoot := common.Hash{}
	signature, err := runPostValidatorsRequest(t, session, validatorDetails, beaconDepositRoot)
	require.NoError(t, err)
	require.NotEmpty(t, signature, "Expected a valid signature from the backend")

	// Verify the new validator count
	metaAfter := runGetValidatorsMetaRequest(t, session)
	require.Equal(t, metaAfter.Active, uint(numValidatorsToRegister))
	require.Equal(t, metaAfter.Max, uint(10)) // Should stay the same

	// Verify
	// GET v3/modules/stakewise/{deployment}/{vault}/validators

	t.Logf("Successfully registered %d validators. New active count: %d",
		numValidatorsToRegister, metaAfter.Active)
}

func runPostValidatorsRequest(t *testing.T, session *db.Session, validatorDetails []stakewise.ValidatorRegistrationDetails, beaconDepositRoot common.Hash) (string, error) {
	// Create the client
	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	response, err := client.StakeWise.Validators_Post(context.Background(), logger, test.Network, test.StakeWiseVaultAddress, validatorDetails, beaconDepositRoot)
	require.NoError(t, err)
	t.Logf("Ran POST /validators request with %d validators", len(validatorDetails))

	return response.Signature, err
}
