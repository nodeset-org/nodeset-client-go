package v0server_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

// Make sure the correct response is returned for a successful request
func TestDepositDataMeta(t *testing.T) {
	depositDataSet := 192

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
	node0Key, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	user, err := db.Core.AddUser(test.User0Email)
	require.NoError(t, err)
	node := user.WhitelistNode(node0Pubkey)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key)
	require.NoError(t, err)
	err = node.Register(regSig)
	require.NoError(t, err)
	vault := deployment.AddStakeWiseVault(test.StakeWiseVaultAddress)
	vault.LatestDepositDataSetIndex = depositDataSet

	// Create a session
	session := db.Core.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node0Pubkey, node0Key)
	require.NoError(t, err)
	err = db.Core.Login(node0Pubkey, session.Nonce, loginSig)
	require.NoError(t, err)

	// Run the request
	data := runDepositDataMetaRequest(t, session)

	// Make sure the response is correct
	require.Equal(t, depositDataSet, data.Version)
	t.Logf("Received correct response - version = %d", data.Version)
}

// Run a GET api/deposit-data/meta request
func runDepositDataMetaRequest(t *testing.T, session *db.Session) stakewise.DepositDataMetaData {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.DepositDataMeta(context.Background(), test.StakeWiseVaultAddress, test.Network)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}
