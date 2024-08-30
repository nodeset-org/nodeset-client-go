package db

import (
	"log/slog"
)

// Mock database for storing nodeset.io info
type Database struct {
	Core          *Database_Core
	Constellation *Database_Constellation
	StakeWise     *Database_StakeWise

	// Internal fields
	logger *slog.Logger
}

// Creates a new database
func NewDatabase(logger *slog.Logger) *Database {
	db := &Database{
		logger: logger,
	}
	db.Core = newDatabase_Core(db, logger)
	db.Constellation = newDatabase_Constellation(logger)
	db.StakeWise = newDatabase_StakeWise(db, logger)
	return db
}

// Clones the database
func (d *Database) Clone() *Database {
	dbClone := &Database{
		logger: d.logger,
	}
	dbClone.Core = d.Core.Clone(dbClone)
	dbClone.Constellation = d.Constellation.Clone()
	dbClone.StakeWise = d.StakeWise.Clone(dbClone)
	return dbClone
}
