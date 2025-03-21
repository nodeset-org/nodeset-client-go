package v3server_stakewise_test

import (
	"fmt"
	"testing"

	stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
	"github.com/rocket-pool/node-manager-core/beacon"

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
	require.Equal(t, 0, metaBefore.Active) // No validators yet
	require.Equal(t, metaBefore.Max, 10)   // Max set to 10
	require.Equal(t, metaBefore.Available, 10)

	// Generate validator details
	numValidatorsToRegister := 3
	validatorDetails := make([]stakewise.ValidatorRegistrationDetails, numValidatorsToRegister)
	for i := 0; i < numValidatorsToRegister; i++ {
		validatorDetails[i] = stakewise.ValidatorRegistrationDetails{
			DepositData: beacon.ExtendedDepositData{},
			ExitMessage: fmt.Sprintf("TODO"),
		}
	}

	// Submit the request (TODO)
	beaconDepositRoot := ""
	signature, err := runPostValidatorsRequest(t, session, validatorDetails, beaconDepositRoot)
	require.NoError(t, err)
	require.NotEmpty(t, signature, "Expected a valid signature from the backend")

	// Verify the new validator count
	metaAfter := runGetValidatorsMetaRequest(t, session)
	require.Equal(t, metaAfter.Active, numValidatorsToRegister)
	require.Equal(t, metaAfter.Max, 10) // Should stay the same
	require.Equal(t, metaAfter.Available, 10-numValidatorsToRegister)

	t.Logf("Successfully registered %d validators. New active count: %d, available count: %d",
		numValidatorsToRegister, metaAfter.Active, metaAfter.Available)
}

func runPostValidatorsRequest(t *testing.T, session *db.Session, validatorDetails []stakewise.ValidatorRegistrationDetails, beaconDepositRoot string) (string, error) {
}
