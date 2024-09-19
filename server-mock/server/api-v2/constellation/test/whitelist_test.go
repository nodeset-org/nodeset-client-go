package v2server_constellation_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	v2core "github.com/nodeset-org/nodeset-client-go/api-v2/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

const (
	whitelist_signature string = "0xf2b73cd729a9b15e8f17ce0189c4ddfe63ad35917f63e2b1ffa7ea1dc527bdf535ba05ba44d2dce733096b8c389472e81a4548b1d75a600633c4ac4bcb8e7c6f1b"
)

func TestGetWhitelist_Unregistered(t *testing.T) {
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
	deployment := db.Constellation.AddDeployment(test.Network, test.ChainIDBig, test.WhitelistAddress, test.SuperNodeAddress)
	node4Key, err := test.GetEthPrivateKey(4)
	require.NoError(t, err)
	node4Pubkey := crypto.PubkeyToAddress(node4Key.PublicKey)
	user, err := db.Core.AddUser(test.User0Email)
	require.NoError(t, err)
	node := user.WhitelistNode(node4Pubkey)
	require.NoError(t, err)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node4Pubkey, node4Key, v2core.NodeAddressMessageFormat)
	require.NoError(t, err)
	err = node.Register(regSig, v2core.NodeAddressMessageFormat)
	require.NoError(t, err)

	// Create a session
	session := db.Core.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node4Pubkey, node4Key)
	require.NoError(t, err)
	err = db.Core.Login(node4Pubkey, session.Nonce, loginSig)
	require.NoError(t, err)

	// Set the admin private key (just the first Hardhat address)
	adminKey, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	deployment.SetAdminPrivateKey(adminKey)

	// Get the registered address
	data := runGetWhitelistRequest(t, session)
	emptyAddress := common.Address{}
	require.False(t, data.Whitelisted)
	require.Equal(t, emptyAddress, data.Address)
	t.Logf("Received correct response - user hasn't whitelisted an address yet")
}

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
	db := mgr.GetDatabase()
	deployment := db.Constellation.AddDeployment(test.Network, test.ChainIDBig, test.WhitelistAddress, test.SuperNodeAddress)
	node4Key, err := test.GetEthPrivateKey(4)
	require.NoError(t, err)
	node4Pubkey := crypto.PubkeyToAddress(node4Key.PublicKey)
	user, err := db.Core.AddUser(test.User0Email)
	require.NoError(t, err)
	node := user.WhitelistNode(node4Pubkey)
	require.NoError(t, err)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node4Pubkey, node4Key, v2core.NodeAddressMessageFormat)
	require.NoError(t, err)
	err = node.Register(regSig, v2core.NodeAddressMessageFormat)
	require.NoError(t, err)

	// Create a session
	session := db.Core.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node4Pubkey, node4Key)
	require.NoError(t, err)
	err = db.Core.Login(node4Pubkey, session.Nonce, loginSig)
	require.NoError(t, err)

	// Set the admin private key (just the first Hardhat address)
	adminKey, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	deployment.SetAdminPrivateKey(adminKey)

	// Create the request
	postData := runPostWhitelistRequest(t, session)

	// Make sure the response is correct
	require.Equal(t, whitelist_signature, postData.Signature)
	t.Logf("Received correct signature response:\nSignature = %s", postData.Signature)

	// Get the registered address
	getData := runGetWhitelistRequest(t, session)
	require.True(t, getData.Whitelisted)
	require.Equal(t, node4Pubkey, getData.Address)
	t.Logf("Received correct registered response - user has whitelisted the correct address")
}

// Run a GET api/v2/modules/constellation/{deployment}/whitelist request
func runGetWhitelistRequest(t *testing.T, session *db.Session) v2constellation.Whitelist_GetData {
	// Create the client
	client := apiv2.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.Constellation.Whitelist_Get(context.Background(), test.Network)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a POST api/v2/modules/constellation/{deployment}/whitelist request
func runPostWhitelistRequest(t *testing.T, session *db.Session) v2constellation.Whitelist_PostData {
	// Create the client
	client := apiv2.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.Constellation.Whitelist_Post(context.Background(), test.Network)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}
