package db

import (
	"log/slog"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// Database for Constellation module info
type Database_Constellation struct {
	// Collection of deployments
	Deployments map[string]*ConstellationDeployment

	// Internal fields
	logger *slog.Logger
}

// Create a new Constellation database
func newDatabase_Constellation(logger *slog.Logger) *Database_Constellation {
	return &Database_Constellation{
		Deployments: map[string]*ConstellationDeployment{},
		logger:      logger,
	}
}

// Clone the database
func (d *Database_Constellation) Clone() *Database_Constellation {
	clone := newDatabase_Constellation(d.logger)
	for id, deployment := range d.Deployments {
		clone.Deployments[id] = deployment.Clone()
	}
	return clone
}

// Adds a deployment - if there is an existing one with the same ID, it will be overwritten to allow for testing changes
func (d *Database_Constellation) AddDeployment(id string, chainID *big.Int, whitelistAddress ethcommon.Address, superNodeAddress ethcommon.Address) *ConstellationDeployment {
	d.Deployments[id] = newConstellationDeployment(id, chainID, whitelistAddress, superNodeAddress)
	return d.Deployments[id]
}

// Gets a deployment by its ID. If there isn't one, returns nil
func (d *Database_Constellation) GetDeployment(deploymentID string) *ConstellationDeployment {
	return d.Deployments[deploymentID]
}
