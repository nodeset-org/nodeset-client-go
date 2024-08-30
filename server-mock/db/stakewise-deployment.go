package db

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// Deployment for StakeWise info
type StakeWiseDeployment struct {
	// The deployment's name / ID
	ID string

	// The Ethereum chain ID this deployment is for
	ChainID *big.Int

	// List of StakeWise vaults
	vaults map[ethcommon.Address]*StakeWiseVault
	db     *Database
}

// Create a new StakeWise deployment
func newStakeWiseDeployment(db *Database, id string, chainID *big.Int) *StakeWiseDeployment {
	return &StakeWiseDeployment{
		ID:      id,
		ChainID: chainID,
		vaults:  make(map[ethcommon.Address]*StakeWiseVault),
		db:      db,
	}
}

// clone the deployment
func (d *StakeWiseDeployment) clone(dbClone *Database) *StakeWiseDeployment {
	clone := newStakeWiseDeployment(dbClone, d.ID, d.ChainID)
	for address, vault := range d.vaults {
		clone.vaults[address] = vault.clone(clone)
	}
	return clone
}

// Add a new StakeWise vault to the deployment. If one already exists with that address, it is just returned.
func (d *StakeWiseDeployment) AddVault(address ethcommon.Address) *StakeWiseVault {
	vault, exists := d.vaults[address]
	if exists {
		return vault
	}
	vault = newStakeWiseVault(d, address)
	d.vaults[address] = vault
	return vault
}

// Get a StakeWise vault by its address. If there isn't one, returns nil
func (d *StakeWiseDeployment) GetVault(address ethcommon.Address) *StakeWiseVault {
	return d.vaults[address]
}

// Get all vaults
func (d *StakeWiseDeployment) GetVaults() map[ethcommon.Address]*StakeWiseVault {
	return d.vaults
}
