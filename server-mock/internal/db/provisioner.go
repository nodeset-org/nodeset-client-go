package db

import (
	"crypto/ecdsa"
	"log/slog"
	"strconv"
	"testing"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/rocket-pool/node-manager-core/node/validator"
	"github.com/stretchr/testify/require"
	types "github.com/wealdtech/go-eth2-types/v2"
)

var (
	NodeKeys   map[uint]*ecdsa.PrivateKey    = map[uint]*ecdsa.PrivateKey{}
	BeaconKeys map[uint]*types.BLSPrivateKey = map[uint]*types.BLSPrivateKey{}
)

// Create a full database for testing
func ProvisionFullDatabase(t *testing.T, logger *slog.Logger, includeDepositDataSet bool) *db.Database {
	// Make the DB
	database := db.NewDatabase(logger)
	swDeployment := database.StakeWise.AddDeployment(test.Network, test.ChainIDBig)
	_ = database.Constellation.AddDeployment(test.Network, test.ChainIDBig, test.WhitelistAddress, test.SuperNodeAddress)

	// Add a StakeWise vault to the database
	vault := swDeployment.AddVault(test.StakeWiseVaultAddress)
	t.Log("Added StakeWise vault to database")

	// Add a users to the database
	_ = addUserToDatabase(t, database, test.User0Email)
	user1 := addUserToDatabase(t, database, test.User1Email)
	user2 := addUserToDatabase(t, database, test.User2Email)
	user3 := addUserToDatabase(t, database, test.User3Email)
	t.Log("Added users to database")

	// Add nodes to the user
	node0 := createNodeAndAddToDatabase(t, database, user1, 0)
	node1 := createNodeAndAddToDatabase(t, database, user2, 1)
	node2 := createNodeAndAddToDatabase(t, database, user3, 2)
	node3 := createNodeAndAddToDatabase(t, database, user3, 3)
	t.Log("Added nodes to users")

	// Get some deposit data
	depositData0 := GenerateDepositData(t, 0, test.StakeWiseVaultAddress)
	depositData1 := GenerateDepositData(t, 1, test.StakeWiseVaultAddress)
	depositData2 := GenerateDepositData(t, 2, test.StakeWiseVaultAddress)
	depositData3 := GenerateDepositData(t, 3, test.StakeWiseVaultAddress)
	depositData4 := GenerateDepositData(t, 4, test.StakeWiseVaultAddress)
	t.Log("Generated deposit data")

	// Handle the deposit data upload
	err := vault.HandleDepositDataUpload(node0, []beacon.ExtendedDepositData{depositData0})
	require.NoError(t, err)
	err = vault.HandleDepositDataUpload(node1, []beacon.ExtendedDepositData{depositData1, depositData2})
	require.NoError(t, err)
	err = vault.HandleDepositDataUpload(node2, []beacon.ExtendedDepositData{depositData3})
	require.NoError(t, err)
	err = vault.HandleDepositDataUpload(node3, []beacon.ExtendedDepositData{depositData4})
	require.NoError(t, err)
	t.Log("Handled deposit data upload")

	// Shortcut if skipping deposit data set generation
	if !includeDepositDataSet {
		return database
	}

	// Create a new set with 1 DD per user and verify
	depositDataSet := vault.CreateNewDepositDataSet(1)
	require.Equal(t, []beacon.ExtendedDepositData{depositData0, depositData1, depositData3}, depositDataSet)

	// Handle the deposit data upload
	vault.UploadDepositDataToStakeWise(depositDataSet)
	t.Log("Uploaded deposit data to StakeWise")

	// Finalize the upload
	vault.MarkDepositDataSetUploaded(depositDataSet)
	t.Log("Marked deposit data set uploaded")

	return database
}

// ==========================
// === Internal Functions ===
// ==========================

// Add a user to the database
func addUserToDatabase(t *testing.T, db *db.Database, userEmail string) *db.User {
	user, err := db.Core.AddUser(userEmail)
	if err != nil {
		t.Fatalf("Error adding user [%s] to database: %v", userEmail, err)
	}
	return user
}

// Create a node, register it with the user, and log it in with a new session
func createNodeAndAddToDatabase(t *testing.T, db *db.Database, user *db.User, index uint) *db.Node {
	nodeKey, exists := NodeKeys[index]
	if !exists {
		var err error
		nodeKey, err = test.GetEthPrivateKey(index)
		if err != nil {
			t.Fatalf("Error getting private key for node 0: %v", err)
		}
		NodeKeys[index] = nodeKey
	}
	nodeAddress := crypto.PubkeyToAddress(nodeKey.PublicKey)

	// Whitelist the node
	node := user.WhitelistNode(nodeAddress)

	// Register the node
	err := node.RegisterWithoutSignature()
	require.NoError(t, err)

	// Create a new session for it
	session := db.Core.CreateSession()
	err = db.Core.LoginWithoutSignature(nodeAddress, session.Nonce)
	if err != nil {
		t.Fatalf("Error logging in node [%s]: %v", nodeAddress.Hex(), err)
	}
	return node
}

// Generate a validator private key and deposit data for the given index
func GenerateDepositData(t *testing.T, index uint, withdrawalAddress ethcommon.Address) beacon.ExtendedDepositData {
	validatorKey, exists := BeaconKeys[index]
	if !exists {
		var err error
		validatorKey, err = test.GetBeaconPrivateKey(index)
		if err != nil {
			t.Fatalf("Error getting private key for validator %d: %v", index, err)
		}
		BeaconKeys[index] = validatorKey
	}
	depositData, err := validator.GetDepositData(
		validatorKey,
		validator.GetWithdrawalCredsFromAddress(withdrawalAddress),
		test.GenesisForkVersion,
		test.DepositAmount,
		test.Network,
	)
	if err != nil {
		t.Fatalf("Error generating deposit data for validator %d: %v", index, err)
	}
	return depositData
}

// Generate a signed exit for the given validator index
func GenerateSignedExit(t *testing.T, index uint) common.ExitData {
	// Create the exit domain
	domain, err := types.ComputeDomain(types.DomainVoluntaryExit, test.CapellaForkVersion, test.GenesisValidatorsRoot)
	if err != nil {
		t.Fatalf("Error computing domain for validator %d: %v", index, err)
	}

	// Get the validator key
	validatorKey, exists := BeaconKeys[index]
	if !exists {
		var err error
		validatorKey, err = test.GetBeaconPrivateKey(index)
		if err != nil {
			t.Fatalf("Error getting private key for validator %d: %v", index, err)
		}
		BeaconKeys[index] = validatorKey
	}

	// Get the exit signature
	validatorIndex := strconv.FormatUint(uint64(index), 10)
	exitSignature, err := validator.GetSignedExitMessage(
		validatorKey,
		validatorIndex,
		test.ExitEpoch,
		domain,
	)
	if err != nil {
		t.Fatalf("Error generating signed exit for validator %d: %v", index, err)
	}

	// Return the exit data
	pubkey := beacon.ValidatorPubkey(validatorKey.PublicKey().Marshal())
	return common.ExitData{
		Pubkey: pubkey.HexWithPrefix(),
		ExitMessage: common.ExitMessage{
			Message: common.ExitMessageDetails{
				Epoch:          strconv.FormatUint(test.ExitEpoch, 10),
				ValidatorIndex: validatorIndex,
			},
			Signature: exitSignature.HexWithPrefix(),
		},
	}
}
