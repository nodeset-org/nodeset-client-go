package v2server_constellation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

const (
	whitelist_signature string = "0x8d6779cdc17bbfd0416fce5af7e4bde2b106ea5904d4c532eee8dfd73e60019b08c35f86b5dd94713b1dc30fa2fc8f91dd1bd32ab2592c22bfade08bfab3817d1b"
)

func TestConstellationWhitelist(t *testing.T) {
	// Take a snapshot
	mgr.TakeSnapshot("test")
	defer func() {
		err := mgr.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	mgr.SetDeployment(test.GetTestDeployment())
	node4Key, err := test.GetEthPrivateKey(4)
	require.NoError(t, err)
	node4Pubkey := crypto.PubkeyToAddress(node4Key.PublicKey)
	err = mgr.AddUser(test.User0Email)
	require.NoError(t, err)
	err = mgr.WhitelistNodeAccount(test.User0Email, node4Pubkey)
	require.NoError(t, err)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node4Pubkey, node4Key)
	require.NoError(t, err)
	err = mgr.RegisterNodeAccount(test.User0Email, node4Pubkey, regSig)
	require.NoError(t, err)

	// Create a session
	session := mgr.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node4Pubkey, node4Key)
	require.NoError(t, err)

	err = mgr.Login(session.Nonce, node4Pubkey, loginSig)
	if err != nil {
		t.Fatalf("error logging in: %v", err)
	}

	// Set the admin private key (just the first Hardhat address)
	adminKey, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	mgr.SetConstellationAdminPrivateKey(adminKey)

	// Create the request
	data := runWhitelistRequest(t, session)

	// Make sure the response is correct
	require.Equal(t, whitelist_signature, data.Signature)
	t.Logf("Received correct response:\nSignature = %s", data.Signature)
}

// Run a GET api/v2/modules/constellation/{deployment}/whitelist request
func runWhitelistRequest(t *testing.T, session *db.Session) v2constellation.WhitelistData {
	// Create the client
	client := apiv2.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.Constellation.Whitelist(context.Background(), test.Network)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}
