package manager

import (
	"fmt"
	"log/slog"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
)

// Mock manager for the nodeset.io service
type NodeSetMockManager struct {
	database *db.Database

	// Internal fields
	snapshots map[string]*db.Database
	logger    *slog.Logger
}

// Used when a deployment is not found
type ErrInvalidDeployment struct {
	DeploymentID string
}

func NewErrInvalidDeployment(deploymentID string) *ErrInvalidDeployment {
	return &ErrInvalidDeployment{DeploymentID: deploymentID}
}

func (e *ErrInvalidDeployment) Error() string {
	return fmt.Sprintf("deployment with ID [%s] not found", e.DeploymentID)
}

// Used when a StakeWise vault is not found
type ErrInvalidVault struct {
	DeploymentID string
	VaultAddress ethcommon.Address
}

func NewErrInvalidVault(deploymentID string, vaultAddress ethcommon.Address) *ErrInvalidVault {
	return &ErrInvalidVault{
		DeploymentID: deploymentID,
		VaultAddress: vaultAddress,
	}
}

func (e *ErrInvalidVault) Error() string {
	return fmt.Sprintf("StakeWise vault with address [%s] in deployment [%s] not found", e.VaultAddress.Hex(), e.DeploymentID)
}

// ===============
// === Manager ===
// ===============

// Creates a new manager
func NewNodeSetMockManager(logger *slog.Logger) *NodeSetMockManager {
	return &NodeSetMockManager{
		database:  db.NewDatabase(logger),
		snapshots: map[string]*db.Database{},
		logger:    logger,
	}
}

// Get the database the manager is currently using
func (m *NodeSetMockManager) GetDatabase() *db.Database {
	return m.database
}

// Set the database for the manager directly if you need to custom provision it
func (m *NodeSetMockManager) SetDatabase(db *db.Database) {
	m.database = db
}

// Take a snapshot of the current database state
func (m *NodeSetMockManager) TakeSnapshot(name string) {
	m.snapshots[name] = m.database.Clone()
	m.logger.Info("Took DB snapshot", "name", name)
}

// Revert to a snapshot of the database state
func (m *NodeSetMockManager) RevertToSnapshot(name string) error {
	snapshot, exists := m.snapshots[name]
	if !exists {
		return fmt.Errorf("snapshot with name [%s] does not exist", name)
	}
	m.database = snapshot.Clone()
	m.logger.Info("Reverted to DB snapshot", "name", name)
	return nil
}
