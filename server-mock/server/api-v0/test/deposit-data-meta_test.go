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
	node0Key, err := test.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	err = mgr.AddUser(test.User0Email)
	if err != nil {
		t.Fatalf("error adding user: %v", err)
	}
	err = mgr.WhitelistNodeAccount(test.User0Email, node0Pubkey)
	if err != nil {
		t.Fatalf("error whitelisting node account: %v", err)
	}
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key)
	if err != nil {
		t.Fatalf("error getting signature for registration: %v", err)
	}
	err = mgr.RegisterNodeAccount(test.User0Email, node0Pubkey, regSig)
	if err != nil {
		t.Fatalf("error registering node account: %v", err)
	}
	err = mgr.AddStakeWiseVault(test.StakeWiseVaultAddress, test.Network)
	if err != nil {
		t.Fatalf("error adding StakeWise vault to database: %v", err)
	}
	vault := mgr.GetStakeWiseVault(test.StakeWiseVaultAddress, test.Network)
	vault.LatestDepositDataSetIndex = depositDataSet

	// Create a session
	session := mgr.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node0Pubkey, node0Key)
	if err != nil {
		t.Fatalf("error getting signature for login: %v", err)
	}
	err = mgr.Login(session.Nonce, node0Pubkey, loginSig)
	if err != nil {
		t.Fatalf("error logging in: %v", err)
	}

	// Run the request
	data := runDepositDataMetaRequest(t, session)

	// Make sure the response is correct
	require.Equal(t, depositDataSet, data.Version)
	t.Logf("Received correct response - version = %d", data.Version)
}

// Run a GET api/deposit-data/meta request
func runDepositDataMetaRequest(t *testing.T, session *db.Session) stakewise.DepositDataMetaData {
	// Create the client
	client := apiv0.NewNodeSetClient(fmt.Sprintf("http://localhost:%d", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.DepositDataMeta(context.Background(), test.StakeWiseVaultAddress, test.Network)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}
