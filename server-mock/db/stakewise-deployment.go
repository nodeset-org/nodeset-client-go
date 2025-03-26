package db

import (
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

const DefaultAvailableValidators = 10

// Deployment for StakeWise info
type StakeWiseDeployment struct {
	// The deployment's name / ID
	ID string

	// The Ethereum chain ID this deployment is for
	ChainID *big.Int

	// List of StakeWise Vaults
	Vaults map[ethcommon.Address]*StakeWiseVault
	db     *Database

	// Validator counts
	ActiveValidators    uint
	MaxValidators       uint
	AvailableValidators uint
}

// Create a new StakeWise deployment
func newStakeWiseDeployment(db *Database, id string, chainID *big.Int) *StakeWiseDeployment {
	return &StakeWiseDeployment{
		ID:                  id,
		ChainID:             chainID,
		Vaults:              make(map[ethcommon.Address]*StakeWiseVault),
		db:                  db,
		ActiveValidators:    0,
		MaxValidators:       0,
		AvailableValidators: DefaultAvailableValidators,
	}
}

// Clone the deployment
func (d *StakeWiseDeployment) clone(dbClone *Database) *StakeWiseDeployment {
	clone := newStakeWiseDeployment(dbClone, d.ID, d.ChainID)
	for address, vault := range d.Vaults {
		clone.Vaults[address] = vault.clone(clone)
	}
	return clone
}

// Add a new StakeWise vault to the deployment. If one already exists with that address, it is just returned.
func (d *StakeWiseDeployment) AddVault(address ethcommon.Address) *StakeWiseVault {
	vault, exists := d.Vaults[address]
	if exists {
		return vault
	}
	vault = newStakeWiseVault(d, address)
	d.Vaults[address] = vault
	return vault
}

// Get a StakeWise vault by its address. If there isn't one, returns nil
func (d *StakeWiseDeployment) GetVault(address ethcommon.Address) *StakeWiseVault {
	return d.Vaults[address]
}

// Get all vaults
func (d *StakeWiseDeployment) GetVaults() map[ethcommon.Address]*StakeWiseVault {
	return d.Vaults
}

// Get all StakeWise validators
func (d *StakeWiseDeployment) GetAllStakeWiseValidators(node *Node) map[ethcommon.Address][]*StakeWiseValidatorInfo {
	vaultInfos := map[ethcommon.Address][]*StakeWiseValidatorInfo{}
	for vaultAddress, vault := range d.Vaults {
		vaultInfo := []*StakeWiseValidatorInfo{}
		nodeValidators := vault.Validators[node.Address]
		for _, validator := range nodeValidators {
			vaultInfo = append(vaultInfo, validator)
		}
		vaultInfos[vaultAddress] = vaultInfo
	}
	return vaultInfos
}
