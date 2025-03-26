package db

import (
	"log/slog"
	"math/big"
)

// Database for StakeWise module info
type Database_StakeWise struct {
	// Collection of Deployments
	Deployments map[string]*StakeWiseDeployment

	// Internal fields
	logger *slog.Logger
	db     *Database
}

// Create a new StakeWise database
func newDatabase_StakeWise(db *Database, logger *slog.Logger) *Database_StakeWise {
	return &Database_StakeWise{
		Deployments: map[string]*StakeWiseDeployment{},
		db:          db,
		logger:      logger,
	}
}

// Clone the database
func (d *Database_StakeWise) clone(dbClone *Database) *Database_StakeWise {
	clone := newDatabase_StakeWise(dbClone, d.logger)
	for id, deployment := range d.Deployments {
		clone.Deployments[id] = deployment.clone(dbClone)
	}
	return clone
}

// Adds a deployment - if there is an existing one with the same ID, it will be overwritten to allow for testing changes
func (d *Database_StakeWise) AddDeployment(id string, chainID *big.Int) *StakeWiseDeployment {
	d.Deployments[id] = newStakeWiseDeployment(d.db, id, chainID)
	return d.Deployments[id]
}

// Gets a deployment by its ID. If there isn't one, returns nil
func (d *Database_StakeWise) GetDeployment(id string) *StakeWiseDeployment {
	return d.Deployments[id]
}

// Get all deployments
func (d *Database_StakeWise) GetDeployments() map[string]*StakeWiseDeployment {
	return d.Deployments
}
