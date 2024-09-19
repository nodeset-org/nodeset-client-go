package db

import (
	"log/slog"
	"testing"

	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/beacon"
	"github.com/stretchr/testify/assert"
)

func TestDatabaseClone(t *testing.T) {
	// Set up a database
	logger := slog.Default()
	db := ProvisionFullDatabase(t, logger, true)

	// Clone the database
	clone := db.Clone()
	t.Log("Cloned database")

	// Check the clone is not the same as the original
	if clone == db {
		t.Fatalf("Clone is the same as the original database")
	}
	compareDatabases(t, db, clone)
	if t.Failed() {
		return
	}
	t.Log("Clone has identical contents to the original database but different pointers")

	// Get the first pubkey from user 2 that hasn't been uploaded yet
	user := db.Core.GetUser(test.User2Email)
	vault := db.StakeWise.GetDeployment(test.Network).GetVault(test.StakeWiseVaultAddress)
	var pubkey beacon.ValidatorPubkey
	found := false
	for _, node := range user.GetNodes() {
		validators := vault.GetStakeWiseValidatorsForNode(node)
		for _, validator := range validators {
			if vault.UploadedData[validator.Pubkey] {
				continue
			}
			pubkey = validator.Pubkey
			found = true
			break
		}
		if found {
			break
		}
	}
	if !found {
		t.Fatalf("Couldn't find a pubkey to test with")
	}
	t.Logf("Using pubkey %s for testing", pubkey.HexWithPrefix())

	// Mark the pubkey as uploaded in the original database
	assert.Equal(t, false, vault.UploadedData[pubkey])
	vault.MarkDepositDataUploaded(pubkey)
	t.Log("Marked deposit data uploaded for StakeWise vault")

	// Make sure the clone didn't get the update
	if clone.StakeWise.GetDeployment(test.Network).GetVault(test.StakeWiseVaultAddress).UploadedData[pubkey] {
		t.Fatalf("Clone got the update")
	}
	t.Log("Clone wasn't updated, as expected")
}

// ==========================
// === Internal Functions ===
// ==========================

// Compare two databases
func compareDatabases(t *testing.T, db *db.Database, clone *db.Database) {
	// Compare StakeWise vault networks
	assert.Equal(t, db.StakeWise.GetDeployments(), clone.StakeWise.GetDeployments())
	for id, deployment := range db.StakeWise.GetDeployments() {
		cloneDeployment := clone.StakeWise.GetDeployment(id)
		for address, vault := range deployment.GetVaults() {
			cloneVault := cloneDeployment.GetVault(address)
			assert.NotSame(t, vault, cloneVault)
		}
	}

	// Compare users
	assert.Equal(t, db.Core.GetUsers(), clone.Core.GetUsers())

	// Make sure the user pointers are all different
	for _, user := range db.Core.GetUsers() {
		cloneUser := clone.Core.GetUser(user.Email)
		assert.NotSame(t, user, cloneUser)
		for nodeAddress, node := range user.GetNodes() {
			cloneNode := cloneUser.GetNode(nodeAddress)
			assert.NotSame(t, node, cloneNode)

			for _, deployment := range db.StakeWise.GetDeployments() {
				cloneDeployment := clone.StakeWise.GetDeployment(deployment.ID)
				for _, vault := range deployment.GetVaults() {
					cloneVault := cloneDeployment.GetVault(vault.Address)

					nodeValidators := vault.GetStakeWiseValidatorsForNode(node)
					cloneValidators := cloneVault.GetStakeWiseValidatorsForNode(cloneNode)
					for pubkey, validator := range nodeValidators {
						cloneValidator := cloneValidators[pubkey]
						assert.NotSame(t, validator, cloneValidator)
					}
				}
			}
		}
	}
}
