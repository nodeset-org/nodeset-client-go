package v2server_constellation_test

import (
	"context"
	"fmt"
	"math/big"
	"testing"

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
	mds_signature string = "0x03de7587ca8f21acfc6654151aded28c5aacbc36de5f30b35fa20c3485f94fff6781355bf7091528376d2fdf01eda7a0e4d75c1995b84dae7a0943c132cfbcf11b"
	mds_salt      string = "90de5e7"
	mds_mpAddress string = "0x21Aa2360e734b11BDE49F2C73d0CF751f4B2a4C3"
)

func TestConstellationDeposit(t *testing.T) {
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

	// Run the request
	salt, _ := big.NewInt(0).SetString(mds_salt, 16)
	data := runMinipoolDepositSignatureRequest(t, session, ethcommon.HexToAddress(mds_mpAddress), salt)

	// Make sure the response is correct
	require.Equal(t, mds_signature, data.Signature)
	t.Logf("Received correct response:\nSignature = %s", data.Signature)
}

// Run a GET api/v2/modules/constellation/{deployment}/minipool/deposit-signature request
func runMinipoolDepositSignatureRequest(t *testing.T, session *db.Session, minipoolAddress ethcommon.Address, salt *big.Int) v2constellation.MinipoolDepositSignatureData {
	// Create the client
	client := apiv2.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.Constellation.MinipoolDepositSignature(context.Background(), test.Network, minipoolAddress, salt)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}
