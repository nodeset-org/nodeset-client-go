package v3server_stakewise_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	apiv3 "github.com/nodeset-org/nodeset-client-go/api-v3"
	v3core "github.com/nodeset-org/nodeset-client-go/api-v3/core"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

// Make sure the correct response is returned for a successful request
func TestGetValidatorsMeta(t *testing.T) {
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
	_ = deployment.AddVault(test.StakeWiseVaultAddress)
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

	// Run a get validators request
	data := runGetValidatorsMetaRequest(t, session)

	// Make sure the response is correct
	require.Equal(t, data.Active, 0)
	require.Equal(t, data.Max, deployment.MaxValidators)
	require.Equal(t, data.Available, deployment.MaxValidators-data.Active)
	t.Logf("Received correct response -  active: %d, max: %d, available: %d", data.Active, data.Max, data.Available)
}

// Run a GET api/validators/meta request
func runGetValidatorsMetaRequest(t *testing.T, session *db.Session) stakewise.VaultsMetaData {
	// Create the client
	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.StakeWise.ValidatorMeta_Get(context.Background(), logger, test.Network, test.StakeWiseVaultAddress)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a PATCH api/validators request
// func runUploadSignedExitsRequest(t *testing.T, session *db.Session, signedExits []common.EncryptedExitData) {
// 	// Create the client
// 	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
// 	client.SetSessionToken(session.Token)

// 	// Run the request
// 	err := client.StakeWise.Validators_Patch(context.Background(), logger, test.Network, test.StakeWiseVaultAddress, signedExits)
// 	require.NoError(t, err)
// 	t.Logf("Ran request")
// }
