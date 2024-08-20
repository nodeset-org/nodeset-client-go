package v2server_constellation_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

const (
	whitelist_timestamp int64  = 1721417393
	whitelist_address   string = "0x1E3b98102e19D3a164d239BdD190913C2F02E756"
	whitelist_chainId   int64  = 31337
	whitelist_signature string = "0xdd45a03d896d93e4fd2ee947bed23fb4f87a24d528cd5ecfe847f4c521cba8c1519f4fbc74d9a12d40fa64244a0616370ae709394a0217659d028351bb8dc3c21b"
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

	// Set the manual timestamp
	manualTime := time.Unix(whitelist_timestamp, 0)
	mgr.SetManualSignatureTimestamp(&manualTime)

	// Create the request
	data := runWhitelistRequest(t, session)

	// Make sure the response is correct
	parsedTime := time.Unix(data.Time, 0)
	require.Equal(t, manualTime, parsedTime)
	require.Equal(t, whitelist_signature, data.Signature)
	t.Logf("Received correct response:\nTime = %s\nSignature = %s", parsedTime, data.Signature)
}

// Run a GET api/v2/modules/constellation/{deployment}/whitelist request
func runWhitelistRequest(t *testing.T, session *db.Session) v2constellation.WhitelistData {
	// Create the client
	client := apiv2.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.Constellation.Whitelist(context.Background(), test.Network, ethcommon.HexToAddress(whitelist_address))
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}
