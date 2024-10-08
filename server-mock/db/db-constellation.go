package db

import (
	"log/slog"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// Database for Constellation module info
type Database_Constellation struct {
	// Collection of deployments
	deployments map[string]*ConstellationDeployment

	// Internal fields
	logger *slog.Logger
	db     *Database
}

// Create a new Constellation database
func newDatabase_Constellation(db *Database, logger *slog.Logger) *Database_Constellation {
	return &Database_Constellation{
		deployments: map[string]*ConstellationDeployment{},
		logger:      logger,
		db:          db,
	}
}

// Clone the database
func (d *Database_Constellation) clone(dbClone *Database) *Database_Constellation {
	clone := newDatabase_Constellation(dbClone, d.logger)
	for id, deployment := range d.deployments {
		clone.deployments[id] = deployment.clone(dbClone)
	}
	return clone
}

// Adds a deployment - if there is an existing one with the same ID, it will be overwritten to allow for testing changes
func (d *Database_Constellation) AddDeployment(id string, chainID *big.Int, whitelistAddress ethcommon.Address, superNodeAddress ethcommon.Address) *ConstellationDeployment {
	d.deployments[id] = newConstellationDeployment(d.db, id, chainID, whitelistAddress, superNodeAddress)
	return d.deployments[id]
}

// Gets a deployment by its ID. If there isn't one, returns nil
func (d *Database_Constellation) GetDeployment(deploymentID string) *ConstellationDeployment {
	return d.deployments[deploymentID]
}

// Get all deployments
func (d *Database_Constellation) GetDeployments() map[string]*ConstellationDeployment {
	return d.deployments
}
