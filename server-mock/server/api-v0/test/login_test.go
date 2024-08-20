package v0server_test

import (
	"context"
	"fmt"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
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
	node0Key, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	err = mgr.AddUser(test.User0Email)
	require.NoError(t, err)
	err = mgr.WhitelistNodeAccount(test.User0Email, node0Pubkey)
	require.NoError(t, err)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key)
	require.NoError(t, err)
	err = mgr.RegisterNodeAccount(test.User0Email, node0Pubkey, regSig)
	require.NoError(t, err)

	// Create a session
	session := mgr.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node0Pubkey, node0Key)
	require.NoError(t, err)

	// Run the request
	data := runLoginRequest(t, session, node0Pubkey, loginSig)

	// Make sure the response is correct
	require.Equal(t, session.Token, data.Token)
	t.Logf("Received correct response - session token = %s", session.Token)
}

// Run a POST api/login request
func runLoginRequest(t *testing.T, session *db.Session, nodeAddress ethcommon.Address, loginSig []byte) core.LoginData {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.Login(context.Background(), session.Nonce, nodeAddress, loginSig)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}
