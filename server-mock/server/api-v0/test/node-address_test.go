package v0server_test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	nsutil "github.com/nodeset-org/nodeset-client-go/utils"
	"github.com/stretchr/testify/require"
)

// Make sure node registration works properly
func TestRegisterNode(t *testing.T) {
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
	//deployment := db.StakeWise.AddDeployment(test.Network, test.ChainIDBig)
	node0Key, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	user, err := db.Core.AddUser(test.User0Email)
	require.NoError(t, err)
	node := user.WhitelistNode(node0Pubkey)

	// Send the request
	runNodeAddressRequest(t, test.User0Email, node0Pubkey, node0Key)
	require.True(t, node.IsRegistered())
	t.Logf("Node registered successfully")
}

// Run a POST api/node-address request
func runNodeAddressRequest(t *testing.T, email string, nodeAddress common.Address, key *ecdsa.PrivateKey) {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)

	// Run the request
	signer := func(message []byte) ([]byte, error) {
		return nsutil.CreateSignature(message, key)
	}
	err := client.NodeAddress(context.Background(), logger, email, nodeAddress, signer)
	require.NoError(t, err)
	t.Logf("Ran request")
}
