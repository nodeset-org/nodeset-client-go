package server

import (
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
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
	server.manager.TakeSnapshot("test")
	defer func() {
		err := server.manager.RevertToSnapshot("test")
		if err != nil {
			t.Fatalf("error reverting to snapshot: %v", err)
		}
	}()

	// Provision the database
	node4Key, err := test.GetEthPrivateKey(4)
	require.NoError(t, err)
	node4Pubkey := crypto.PubkeyToAddress(node4Key.PublicKey)
	err = server.manager.AddUser(test.User0Email)
	require.NoError(t, err)
	err = server.manager.WhitelistNodeAccount(test.User0Email, node4Pubkey)
	require.NoError(t, err)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node4Pubkey, node4Key)
	require.NoError(t, err)
	err = server.manager.RegisterNodeAccount(test.User0Email, node4Pubkey, regSig)
	require.NoError(t, err)

	// Create a session
	session := server.manager.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node4Pubkey, node4Key)
	require.NoError(t, err)

	err = server.manager.Login(session.Nonce, node4Pubkey, loginSig)
	if err != nil {
		t.Fatalf("error logging in: %v", err)
	}

	// Set the admin private key (just the first Hardhat address)
	adminKey, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	server.manager.SetConstellationAdminPrivateKey(adminKey)

	// Set the manual timestamp
	manualTime := time.Unix(whitelist_timestamp, 0)
	server.manager.SetManualSignatureTimestamp(&manualTime)

	// Create the request
	chainID := big.NewInt(whitelist_chainId)
	whitelistAddress := ethcommon.HexToAddress(whitelist_address)
	request, err := http.NewRequest(http.MethodGet, fmt.Sprintf("http://localhost:%d/api/v2/modules/constellation/%s", port, v2constellation.WhitelistPath), nil)
	if err != nil {
		t.Fatalf("error creating request: %v", err)
	}
	query := request.URL.Query()
	query.Add("chainId", chainID.String())
	query.Add("whitelistAddress", whitelistAddress.Hex())
	request.URL.RawQuery = query.Encode()
	t.Logf("Created request")

	// Add the auth header
	auth.AddAuthorizationHeader(request, session)
	t.Logf("Added auth header")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		t.Fatalf("error sending request: %v", err)
	}
	t.Logf("Sent request")

	// Check the status code
	require.Equal(t, http.StatusOK, response.StatusCode)
	t.Logf("Received OK status code")

	// Read the body
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("error reading the response body: %v", err)
	}
	var parsedResponse common.NodeSetResponse[v2constellation.WhitelistData]
	err = json.Unmarshal(bytes, &parsedResponse)
	if err != nil {
		t.Fatalf("error deserializing response: %v", err)
	}

	// Make sure the response is correct
	parsedTime := time.Unix(parsedResponse.Data.Time, 0)
	require.Equal(t, manualTime, parsedTime)
	require.Equal(t, whitelist_signature, parsedResponse.Data.Signature)
	t.Logf("Received correct response:\nTime = %s\nSignature = %s", parsedTime, parsedResponse.Data.Signature)
}
