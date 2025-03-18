package v3server_constellation_test

import (
	"context"
	"errors"
	"fmt"
	"math/big"
	"testing"

	"filippo.io/age"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	apiv3 "github.com/nodeset-org/nodeset-client-go/api-v3"
	v3constellation "github.com/nodeset-org/nodeset-client-go/api-v3/constellation"
	v3core "github.com/nodeset-org/nodeset-client-go/api-v3/core"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/stretchr/testify/require"
)

func TestGetValidators_Empty(t *testing.T) {
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
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node4Pubkey, node4Key, v3core.NodeAddressMessageFormat)
	require.NoError(t, err)
	err = node.Register(regSig, v3core.NodeAddressMessageFormat)
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

	// Whitelist the node
	runPostWhitelistRequest(t, session)

	// Run the get request
	data := runGetValidatorsRequest(t, session)
	require.Empty(t, data.Validators)
	t.Logf("Received correct response - validators is empty")
}

func TestPatchValidators(t *testing.T) {
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
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node4Pubkey, node4Key, v3core.NodeAddressMessageFormat)
	require.NoError(t, err)
	err = node.Register(regSig, v3core.NodeAddressMessageFormat)
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

	// Set the encryption identity
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	db.SetSecretEncryptionIdentity(id)

	// Whitelist the node
	runPostWhitelistRequest(t, session)

	// More provisioning
	numValidators := 3
	pubkeys := make([]beacon.ValidatorPubkey, numValidators)
	expectedValidators := map[beacon.ValidatorPubkey]v3constellation.ValidatorStatus{}
	for i := 0; i < numValidators; i++ {
		mpAddress := ethcommon.HexToAddress(fmt.Sprintf("0x90de%d", i))
		pubkey := pubkeys[i]
		pubkey[0] = byte(0xbe)
		pubkey[1] = byte(0xac)
		pubkey[2] = byte(0x09)
		pubkey[3] = byte(i)
		pubkeys[i] = pubkey
		salt := big.NewInt(int64(i))
		runMinipoolDepositSignatureRequest(t, session, mpAddress, salt)
		deployment.SetValidatorInfoForMinipool(mpAddress, pubkey)
		expectedValidators[pubkey] = v3constellation.ValidatorStatus{
			Pubkey:              pubkey,
			RequiresExitMessage: true,
		}
		deployment.IncrementSuperNodeNonce(node.Address)
	}

	// Run the get request
	data := runGetValidatorsRequest(t, session)
	validatorsMap := map[beacon.ValidatorPubkey]v3constellation.ValidatorStatus{}
	for _, validator := range data.Validators {
		validatorsMap[validator.Pubkey] = validator
	}
	for _, validator := range data.Validators {
		validatorsMap[validator.Pubkey] = v3constellation.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			RequiresExitMessage: true,
		}
	}

	// Make sure the response is correct
	require.Equal(t, expectedValidators, validatorsMap)
	t.Logf("Received correct response:\n%d validators set", len(data.Validators))

	// Upload signed exits and verify round trip
	epoch := 12
	for i := 0; i < numValidators; i++ {
		// Create the exit message
		exitMessage := common.ExitMessage{
			Message: common.ExitMessageDetails{
				Epoch:          fmt.Sprintf("%d", epoch),
				ValidatorIndex: fmt.Sprintf("%d", i),
			},
			Signature: fmt.Sprintf("0x%x", i),
		}

		// Encrypt it
		pubkey := id.Recipient().String()
		encryptedMessage, err := common.EncryptSignedExitMessage(exitMessage, pubkey)
		require.NoError(t, err)

		// Run the patch request for each validator
		runPatchValidatorsRequest(t, session, []common.EncryptedExitData{
			{
				Pubkey:      pubkeys[i].Hex(),
				ExitMessage: encryptedMessage,
			},
		})

		// Make sure the response is correct
		data := runGetValidatorsRequest(t, session)
		validatorsMap := map[beacon.ValidatorPubkey]v3constellation.ValidatorStatus{}
		for _, validator := range data.Validators {
			validatorsMap[validator.Pubkey] = validator
		}
		for j := 0; j < numValidators; j++ {
			pubkey := pubkeys[j]
			validator, exists := validatorsMap[pubkey]
			require.True(t, exists)
			if j <= i {
				require.False(t, validator.RequiresExitMessage)
			} else {
				require.True(t, validator.RequiresExitMessage)
			}
		}
	}
	t.Logf("All validator exits uploaded correctly")
}

func TestSignedExitAlreadyExists(t *testing.T) {
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
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node4Pubkey, node4Key, v3core.NodeAddressMessageFormat)
	require.NoError(t, err)
	err = node.Register(regSig, v3core.NodeAddressMessageFormat)
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

	// Set the encryption identity
	id, err := age.GenerateX25519Identity()
	require.NoError(t, err)
	db.SetSecretEncryptionIdentity(id)

	// Whitelist the node
	runPostWhitelistRequest(t, session)

	// More provisioning
	expectedValidators := map[beacon.ValidatorPubkey]v3constellation.ValidatorStatus{}
	mpAddress := ethcommon.HexToAddress("0x90de00")
	pubkey := beacon.ValidatorPubkey{}
	pubkey[0] = byte(0xbe)
	pubkey[1] = byte(0xac)
	pubkey[2] = byte(0x09)
	pubkey[3] = byte(0x00)
	salt := big.NewInt(0)
	runMinipoolDepositSignatureRequest(t, session, mpAddress, salt)
	deployment.SetValidatorInfoForMinipool(mpAddress, pubkey)
	expectedValidators[pubkey] = v3constellation.ValidatorStatus{
		Pubkey:              pubkey,
		RequiresExitMessage: true,
	}
	deployment.IncrementSuperNodeNonce(node.Address)

	// Upload signed exit
	epoch := 12
	// Create the exit message
	exitMessage := common.ExitMessage{
		Message: common.ExitMessageDetails{
			Epoch:          fmt.Sprintf("%d", epoch),
			ValidatorIndex: "0",
		},
		Signature: "0",
	}

	// Encrypt it
	recipientPubkey := id.Recipient().String()
	encryptedMessage, err := common.EncryptSignedExitMessage(exitMessage, recipientPubkey)
	require.NoError(t, err)

	// Run the patch request for each validator
	exitData := []common.EncryptedExitData{
		{
			Pubkey:      pubkey.Hex(),
			ExitMessage: encryptedMessage,
		},
	}
	runPatchValidatorsRequest(t, session, exitData)
	t.Logf("Initial validator exit uploaded successfully")

	// Run the patch request again and check the error code
	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	err = client.Constellation.Validators_Patch(context.Background(), logger, test.Network, exitData)
	require.Error(t, err)
	require.True(t, errors.Is(err, v3constellation.ErrExitMessageExists))
	t.Logf("Received correct exit-message-exists error code")
}

// Run a GET api/v2/modules/constellation/{deployment}/validators request
func runGetValidatorsRequest(t *testing.T, session *db.Session) v3constellation.ValidatorsData {
	// Create the client
	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.Constellation.Validators_Get(context.Background(), logger, test.Network)
	require.NoError(t, err)
	t.Logf("Ran request")
	return data
}

// Run a PATCH api/v2/modules/constellation/{deployment}/validators request
func runPatchValidatorsRequest(t *testing.T, session *db.Session, exitData []common.EncryptedExitData) {
	// Create the client
	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	err := client.Constellation.Validators_Patch(context.Background(), logger, test.Network, exitData)
	require.NoError(t, err)
	t.Logf("Ran request")
}
