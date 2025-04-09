package v3server_stakewise_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	apiv3 "github.com/nodeset-org/nodeset-client-go/api-v3"
	v3core "github.com/nodeset-org/nodeset-client-go/api-v3/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

func TestGetDeployments(t *testing.T) {
	// Take snapshot
	mgr.TakeSnapshot("test")
	defer func() {
		err := mgr.RevertToSnapshot("test")
		require.NoError(t, err)
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

	// Run request
	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)
	resp, err := client.StakeWise.Deployments(context.Background(), logger)
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.Len(t, resp.Deployments, 1)
	require.Equal(t, deployment.ID, resp.Deployments[0].Name)
	require.Equal(t, deployment.ChainID.String(), resp.Deployments[0].ChainID)

	t.Logf("Successfully fetched deployment: %s", resp.Deployments[0].Name)
}
