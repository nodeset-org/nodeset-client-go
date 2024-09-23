package v2server_core_test

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	v2core "github.com/nodeset-org/nodeset-client-go/api-v2/core"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	nsutil "github.com/nodeset-org/nodeset-client-go/utils"
	"github.com/stretchr/testify/require"
)

// Make sure logging in works properly
func TestLogin(t *testing.T) {
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
	node0Key, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	user, err := db.Core.AddUser(test.User0Email)
	require.NoError(t, err)
	node := user.WhitelistNode(node0Pubkey)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key, v2core.NodeAddressMessageFormat)
	require.NoError(t, err)
	err = node.Register(regSig, v2core.NodeAddressMessageFormat)
	require.NoError(t, err)

	// Create a session
	session := db.Core.CreateSession()

	// Run the request
	data := runLoginRequest(t, session, node0Pubkey, node0Key)

	// Make sure the response is correct
	require.Equal(t, session.Token, data.Token)
	t.Logf("Received correct response - session token = %s", session.Token)
}

// Run a POST api/login request
func runLoginRequest(t *testing.T, session *db.Session, nodeAddress ethcommon.Address, key *ecdsa.PrivateKey) core.LoginData {
	// Create the client
	client := apiv2.NewNodeSetClient(logger, fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	signer := func(message []byte) ([]byte, error) {
		return nsutil.CreateSignature(message, key)
	}
	data, err := client.Core.Login(context.Background(), session.Nonce, nodeAddress, signer)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}
