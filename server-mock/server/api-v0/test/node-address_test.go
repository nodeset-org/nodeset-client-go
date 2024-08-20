package v0server_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
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
	node0Key, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	err = mgr.AddUser(test.User0Email)
	require.NoError(t, err)
	err = mgr.WhitelistNodeAccount(test.User0Email, node0Pubkey)
	require.NoError(t, err)

	// Send the request
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key)
	runNodeAddressRequest(t, test.User0Email, node0Pubkey, regSig)
	t.Logf("Node registered successfully")
}

// Run a POST api/node-address request
func runNodeAddressRequest(t *testing.T, email string, nodeAddress common.Address, signature []byte) {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d", port), timeout)

	// Run the request
	err := client.NodeAddress(context.Background(), email, nodeAddress, signature)
	require.NoError(t, err)
	t.Logf("Ran request")
}
