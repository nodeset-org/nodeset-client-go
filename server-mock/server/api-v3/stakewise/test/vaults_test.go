package v3server_stakewise_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	apiv3 "github.com/nodeset-org/nodeset-client-go/api-v3"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

func TestGetVaults(t *testing.T) {
	// Take a snapshot
	mgr.TakeSnapshot("test")
	defer func() {
		err := mgr.RevertToSnapshot("test")
		require.NoError(t, err)
	}()

	// Provision the database
	db := mgr.GetDatabase()
	deployment := db.StakeWise.AddDeployment(test.Network, test.ChainIDBig)
	deployment.AddVault(test.StakeWiseVaultName, test.StakeWiseVaultAddress)

	nodeKey, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	nodePubkey := crypto.PubkeyToAddress(nodeKey.PublicKey)

	user, err := db.Core.AddUser(test.User0Email)
	require.NoError(t, err)

	node := user.WhitelistNode(nodePubkey)

	regSig, err := auth.GetSignatureForRegistration(test.User0Email, nodePubkey, nodeKey, "address")
	require.NoError(t, err)
	require.NoError(t, node.Register(regSig, "address"))

	session := db.Core.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, nodePubkey, nodeKey)
	require.NoError(t, err)
	require.NoError(t, db.Core.Login(nodePubkey, session.Nonce, loginSig))

	// Create client
	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	vaults, err := client.StakeWise.Vaults(context.Background(), logger, test.Network)
	require.NoError(t, err)
	require.Len(t, vaults.Vaults, 1)
	require.Equal(t, vaults.Vaults[0].Name, test.StakeWiseVaultName)
	require.Equal(t, vaults.Vaults[0].Address, test.StakeWiseVaultAddress)

	t.Logf("Successfully fetched %d vault(s)", len(vaults.Vaults))
}
