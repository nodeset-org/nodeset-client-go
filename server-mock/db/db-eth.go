package db

import (
	"log/slog"

	ethcommon "github.com/ethereum/go-ethereum/common"
)

// Database for Ethereum chain mock info
// TEMPORARY until Hardhat support is added
type Database_Ethereum struct {
	// The current deposit root for the Beacon deposit contract
	depositRoot ethcommon.Hash

	// Internal fields
	logger *slog.Logger
	db     *Database
}

// Create a new Ethereum database
func newDatabase_Ethereum(db *Database, logger *slog.Logger) *Database_Ethereum {
	return &Database_Ethereum{
		depositRoot: ethcommon.Hash{},
		logger:      logger,
		db:          db,
	}
}

// Clone the database
func (d *Database_Ethereum) clone(dbClone *Database) *Database_Ethereum {
	clone := newDatabase_Ethereum(dbClone, d.logger)
	clone.depositRoot = d.depositRoot
	return clone
}

// Gets a deployment by its ID. If there isn't one, returns nil
func (d *Database_Ethereum) GetDepositRoot() ethcommon.Hash {
	return d.depositRoot
}

// Sets the Beacon deposit contract's deposit root
func (d *Database_Ethereum) SetDepositRoot(root ethcommon.Hash) {
	d.depositRoot = root
}
