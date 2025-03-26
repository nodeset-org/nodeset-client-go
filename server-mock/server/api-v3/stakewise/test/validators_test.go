package v3server_stakewise_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	apiv3 "github.com/nodeset-org/nodeset-client-go/api-v3"
	stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	v3core "github.com/nodeset-org/nodeset-client-go/api-v3/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/stretchr/testify/require"
)

func TestPostValidators(t *testing.T) {
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
	deployment := db.StakeWise.AddDeployment(test.Network, test.ChainIDBig)
	deployment.AddVault(test.StakeWiseVaultAddress)
	deployment.MaxValidators = 10 // Set max validators

	node0Key, err := test.GetEthPrivateKey(0)
	require.NoError(t, err)
	node0Pubkey := crypto.PubkeyToAddress(node0Key.PublicKey)
	user, err := db.Core.AddUser(test.User0Email)
	require.NoError(t, err)
	node := user.WhitelistNode(node0Pubkey)
	regSig, err := auth.GetSignatureForRegistration(test.User0Email, node0Pubkey, node0Key, v3core.NodeAddressMessageFormat)
	require.NoError(t, err)
	err = node.Register(regSig, v3core.NodeAddressMessageFormat)
	require.NoError(t, err)

	numValidatorsToRegister := 3

	// Create a session
	session := db.Core.CreateSession()
	loginSig, err := auth.GetSignatureForLogin(session.Nonce, node0Pubkey, node0Key)
	require.NoError(t, err)
	err = db.Core.Login(node0Pubkey, session.Nonce, loginSig)
	require.NoError(t, err)

	// Get the initial validator list and check that it's empty
	fetchedValidatorsBefore := runGetValidatorsRequest(t, session)
	require.Len(t, fetchedValidatorsBefore.Validators, 0)

	// Get the initial validator limits
	metaBefore := runGetValidatorsMetaRequest(t, session)
	require.Equal(t, metaBefore.Active, uint(0)) // No validators yet
	require.Equal(t, metaBefore.Max, uint(10))   // Max set to 10
	require.Equal(t, metaBefore.Available, uint(10))

	// Generate validator details
	validatorDetails := make([]stakewise.ValidatorRegistrationDetails, numValidatorsToRegister)
	for i := 0; i < numValidatorsToRegister; i++ {
		pubkey := make([]byte, 48)
		pubkey[0] = byte(i + 1) // Ensure uniqueness

		signature := make([]byte, 96)
		signature[0] = byte(i + 1) // Optional uniqueness

		exitMessage := common.ExitMessage{
			Message: common.ExitMessageDetails{
				Epoch:          fmt.Sprintf("epoch_%d", i),
				ValidatorIndex: fmt.Sprintf("validator_index_%d", i),
			},
		}
		exitMessageBytes, err := json.Marshal(exitMessage)
		require.NoError(t, err)

		validatorDetails[i] = stakewise.ValidatorRegistrationDetails{
			DepositData: beacon.ExtendedDepositData{
				PublicKey: pubkey,
				Signature: signature,
			},
			ExitMessage: string(exitMessageBytes),
		}
	}

	// Submit the request
	beaconDepositRoot := ethcommon.Hash{}
	signature, err := runPostValidatorsRequest(t, session, validatorDetails, beaconDepositRoot)
	require.NoError(t, err)
	require.NotEmpty(t, signature, "Expected a valid signature from the backend")

	// Verify the new validator count
	metaAfter := runGetValidatorsMetaRequest(t, session)
	require.Equal(t, metaAfter.Active, uint(numValidatorsToRegister))
	require.Equal(t, metaAfter.Max, uint(10))      // Should stay the same
	require.Equal(t, metaAfter.Available, uint(7)) // 10 - 3

	// Verify
	// GET v3/modules/stakewise/{deployment}/{vault}/validators
	fetchedValidatorsAfter := runGetValidatorsRequest(t, session)

	expectedPubkeys := make(map[string]bool)
	for _, detail := range validatorDetails {
		expectedPubkeys[beacon.ValidatorPubkey([48]byte(detail.DepositData.PublicKey)).Hex()] = true
	}

	// length check to ensure all validators were registered
	require.Len(t, fetchedValidatorsAfter.Validators, numValidatorsToRegister)
	for _, validator := range fetchedValidatorsAfter.Validators {
		require.True(t, expectedPubkeys[validator.Pubkey.Hex()])
		require.True(t, validator.ExitMessageUploaded)
	}

	t.Logf("Successfully registered %d validators. New active count: %d",
		numValidatorsToRegister, metaAfter.Active)
}

func runPostValidatorsRequest(t *testing.T, session *db.Session, validatorDetails []stakewise.ValidatorRegistrationDetails, beaconDepositRoot ethcommon.Hash) (string, error) {
	// Create the client
	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	response, err := client.StakeWise.Validators_Post(
		context.Background(),
		logger,
		test.Network,
		test.StakeWiseVaultAddress,
		validatorDetails,
		beaconDepositRoot,
	)
	require.NoError(t, err)
	t.Logf("Ran POST /validators request with %d validators", len(validatorDetails))

	return response.Signature, err
}

func runGetValidatorsRequest(t *testing.T, session *db.Session) stakewise.ValidatorsData {
	// Create the client
	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
	client.SetSessionToken(session.Token)

	// Run the request
	data, err := client.StakeWise.Validators_Get(context.Background(), logger, test.Network, test.StakeWiseVaultAddress)
	require.NoError(t, err)
	t.Logf("Ran GET /validators request")
	return data
}
