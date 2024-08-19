package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"testing"
	"time"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

const (
	mds_timestamp        int64  = 1722623396
	mds_supernodeAddress string = "0x8ac5eE52F70AE01dB914bE459D8B3d50126fd6aE"
	mds_chainId          int64  = 31337
	mds_signature        string = "0x0a72f8a916178dd4e01c48869c1a2b08ca052620950c3b7f603500c9a25fa64f1df114e568da21e8ec32b3ea84edc248effcec7efdb78ce2a0b04b8b6cd2624b1c"
	mds_salt             string = "90de5e7"
	mds_mpAddress        string = "0x21Aa2360e734b11BDE49F2C73d0CF751f4B2a4C3"
)

func TestConstellationDeposit(t *testing.T) {
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
	manualTime := time.Unix(mds_timestamp, 0)
	server.manager.SetManualSignatureTimestamp(&manualTime)

	// Create the request
	salt, _ := big.NewInt(0).SetString(mds_salt, 16)
	loginReq := apiv2.MinipoolDepositSignatureRequest{
		MinipoolAddress:  ethcommon.HexToAddress(mds_mpAddress),
		Salt:             salt.String(),
		SuperNodeAddress: ethcommon.HexToAddress(mds_supernodeAddress),
		ChainId:          big.NewInt(mds_chainId).String(),
	}
	body, err := json.Marshal(loginReq)
	require.NoError(t, err)
	request, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://localhost:%d/api/v2/modules/constellation/%s", port, apiv2.MinipoolDepositSignaturePath), bytes.NewReader(body))
	require.NoError(t, err)
	t.Logf("Created request")

	// Add the auth header
	auth.AddAuthorizationHeader(request, session)
	t.Logf("Added auth header")

	// Send the request
	response, err := http.DefaultClient.Do(request)
	require.NoError(t, err)
	t.Logf("Sent request")

	// Read the body
	defer response.Body.Close()
	bytes, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	var parsedResponse common.NodeSetResponse[apiv2.MinipoolDepositSignatureData]
	err = json.Unmarshal(bytes, &parsedResponse)
	require.NoError(t, err)

	// Check the status code
	require.Equal(t, http.StatusOK, response.StatusCode)
	t.Logf("Received OK status code")

	// Make sure the response is correct
	parsedTime := time.Unix(parsedResponse.Data.Time, 0)
	require.Equal(t, manualTime, parsedTime)
	require.Equal(t, mds_signature, parsedResponse.Data.Signature)
	t.Logf("Received correct response:\nTime = %s\nSignature = %s", parsedTime, parsedResponse.Data.Signature)
}
