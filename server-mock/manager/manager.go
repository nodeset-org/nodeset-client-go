package manager

import (
	"crypto/ecdsa"
	"errors"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/utils"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Mock manager for the nodeset.io service
type NodeSetMockManager struct {
	database *db.Database

	// Internal fields
	snapshots map[string]*db.Database
	logger    *slog.Logger
}

var (
	ErrInvalidSession error = errors.New("session token is invalid")
)

// Creates a new manager
func NewNodeSetMockManager(logger *slog.Logger) *NodeSetMockManager {
	return &NodeSetMockManager{
		database:  db.NewDatabase(logger),
		snapshots: map[string]*db.Database{},
		logger:    logger,
	}
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
	m.database = snapshot
	m.logger.Info("Reverted to DB snapshot", "name", name)
	return nil
}

// ================
// === Database ===
// ================

// Sets the deployment for the database
func (m *NodeSetMockManager) SetDeployment(deployment *db.Deployment) {
	m.database.SetDeployment(deployment)
}

// Gets the deployment with the given ID
func (m *NodeSetMockManager) GetDeployment(id string) *db.Deployment {
	if m.database.Deployment == nil || m.database.Deployment.DeploymentID != id {
		return nil
	}
	return m.database.Deployment
}

// Adds a StakeWise vault
func (m *NodeSetMockManager) AddStakeWiseVault(deployment string, address ethcommon.Address) error {
	return m.database.AddStakeWiseVault(deployment, address)
}

// Gets a StakeWise vault
func (m *NodeSetMockManager) GetStakeWiseVault(deployment string, address ethcommon.Address) *db.StakeWiseVault {
	return m.database.GetStakeWiseVault(deployment, address)
}

// Get all of the StakeWise vaults for a deployment
func (m *NodeSetMockManager) GetStakeWiseVaults(deployment string) []*db.StakeWiseVault {
	return m.database.StakeWiseVaults[deployment]
}

// Adds a user to the database
func (m *NodeSetMockManager) AddUser(email string) error {
	return m.database.AddUser(email)
}

// Whitelists a node with a user
func (m *NodeSetMockManager) WhitelistNodeAccount(email string, nodeAddress ethcommon.Address) error {
	return m.database.WhitelistNodeAccount(email, nodeAddress)
}

// Registers a whitelisted node with a user
func (m *NodeSetMockManager) RegisterNodeAccount(email string, nodeAddress ethcommon.Address, signature []byte) error {
	// Verify the signature
	err := auth.VerifyRegistrationSignature(email, nodeAddress, signature)
	if err != nil {
		return err
	}

	// Try to register the node
	return m.database.RegisterNodeAccount(email, nodeAddress)
}

// Creates a new session and returns the nonce for it
func (m *NodeSetMockManager) CreateSession() *db.Session {
	return m.database.CreateSession()
}

// Logs a session in
func (m *NodeSetMockManager) Login(nonce string, nodeAddress ethcommon.Address, signature []byte) error {
	// Verify the signature
	err := auth.VerifyLoginSignature(nonce, nodeAddress, signature)
	if err != nil {
		return err
	}

	// Log the session in
	return m.database.Login(nodeAddress, nonce)
}

// Gets a session by nonce
func (m *NodeSetMockManager) GetSessionByNonce(nonce string) *db.Session {
	return m.database.GetSessionByNonce(nonce)
}

// Gets a session by token
func (m *NodeSetMockManager) GetSessionByToken(token string) *db.Session {
	return m.database.GetSessionByToken(token)
}

// Verifies a request's session and returns the node address the session belongs to
func (m *NodeSetMockManager) VerifyRequest(r *http.Request) (*db.Session, error) {
	token, err := auth.GetSessionTokenFromRequest(r)
	if err != nil {
		return nil, err
	}

	// Get the session
	session := m.database.GetSessionByToken(token)
	if session == nil {
		return nil, ErrInvalidSession
	}
	return session, nil
}

// Get a node by address - returns true if registered, false if just whitelisted
func (m *NodeSetMockManager) GetNode(address ethcommon.Address) (*db.Node, bool) {
	return m.database.GetNode(address)
}

// Get the StakeWise status of a validator
func (m *NodeSetMockManager) GetValidatorStatus(network string, pubkey beacon.ValidatorPubkey) stakewise.StakeWiseStatus {
	vaults, exists := m.database.StakeWiseVaults[network]
	if !exists {
		return stakewise.StakeWiseStatus_Pending
	}

	// Get the validator for this pubkey
	var validator *db.Validator
	for _, user := range m.database.Users {
		for _, node := range user.RegisteredNodes {
			validators, exists := node.Validators[network]
			if !exists {
				continue
			}
			for _, candidate := range validators {
				if candidate.Pubkey == pubkey {
					validator = candidate
					break
				}
			}
		}
		if validator != nil {
			break
		}
	}
	if validator == nil {
		return stakewise.StakeWiseStatus_Pending
	}

	// Check if the StakeWise vault has already seen it
	for _, vault := range vaults {
		if vault.Address == validator.VaultAddress && vault.UploadedData[validator.Pubkey] {
			if validator.MarkedActive {
				return stakewise.StakeWiseStatus_Registered
			}
		}
	}

	// Check to see if the deposit data has been used
	if validator.DepositDataUsed {
		return stakewise.StakeWiseStatus_Uploaded
	}
	return stakewise.StakeWiseStatus_Pending
}

// Handle a new collection of deposit data uploads from a node
func (m *NodeSetMockManager) HandleDepositDataUpload(nodeAddress ethcommon.Address, deployment string, vaultAddress ethcommon.Address, data []beacon.ExtendedDepositData) error {
	return m.database.HandleDepositDataUpload(nodeAddress, deployment, vaultAddress, data)
}

// Handle a new collection of signed exits from a node
func (m *NodeSetMockManager) HandleSignedExitUpload(nodeAddress ethcommon.Address, deployment string, vaultAddress ethcommon.Address, data []common.ExitData) error {
	return m.database.HandleSignedExitUpload(nodeAddress, deployment, vaultAddress, data)
}

// Create a new deposit data set
func (m *NodeSetMockManager) CreateNewDepositDataSet(deployment string, validatorsPerUser int) []beacon.ExtendedDepositData {
	return m.database.CreateNewDepositDataSet(deployment, validatorsPerUser)
}

// Call this to "upload" a deposit data set to StakeWise
func (m *NodeSetMockManager) UploadDepositDataToStakeWise(deployment string, vaultAddress ethcommon.Address, data []beacon.ExtendedDepositData) error {
	return m.database.UploadDepositDataToStakeWise(deployment, vaultAddress, data)
}

// Call this once a deposit data set has been "uploaded" to StakeWise
func (m *NodeSetMockManager) MarkDepositDataSetUploaded(deployment string, vaultAddress ethcommon.Address, data []beacon.ExtendedDepositData) error {
	return m.database.MarkDepositDataSetUploaded(deployment, vaultAddress, data)
}

// Call this once a deposit data set has been "registered" to StakeWise
func (m *NodeSetMockManager) MarkValidatorsRegistered(deployment string, vaultAddress ethcommon.Address, data []beacon.ExtendedDepositData) error {
	return m.database.MarkValidatorsRegistered(deployment, vaultAddress, data)
}

// Call this to set the private key for the ConstellationAdmin contract
func (m *NodeSetMockManager) SetConstellationAdminPrivateKey(privateKey *ecdsa.PrivateKey) {
	m.database.SetConstellationAdminPrivateKey(privateKey)
}

// Call this to get a signature for adding the node to the Constellation whitelist
func (m *NodeSetMockManager) GetConstellationWhitelistSignature(nodeAddress ethcommon.Address, chainId *big.Int, whitelistAddress ethcommon.Address) ([]byte, error) {
	if m.database.ConstellationAdminPrivateKey == nil {
		return nil, fmt.Errorf("constellation admin private key not set")
	}

	chainIdBytes := [32]byte{}
	chainId.FillBytes(chainIdBytes[:])

	nonceBytes := [32]byte{}
	nonce := m.GetWhitelistNonce(nodeAddress)
	nonce.FillBytes(nonceBytes[:])

	sigTypeBytes := [32]byte{} // Always 0 for the mock

	message := crypto.Keccak256(
		nodeAddress[:],
		whitelistAddress[:],
		nonceBytes[:],
		sigTypeBytes[:],
		chainIdBytes[:],
	)

	// Hash of the concatenated addresses
	signature, err := utils.CreateSignature(message, m.database.ConstellationAdminPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating signature: %w", err)
	}
	return signature, nil
}

// Call this to get a signature for depositing a new minipool with Constellation
func (m *NodeSetMockManager) GetConstellationDepositSignature(nodeAddress ethcommon.Address, minipoolAddress ethcommon.Address, salt *big.Int, superNodeAddress ethcommon.Address, chainId *big.Int) ([]byte, error) {
	if m.database.ConstellationAdminPrivateKey == nil {
		return nil, fmt.Errorf("constellation admin private key not set")
	}

	chainIdBytes := [32]byte{}
	chainId.FillBytes(chainIdBytes[:])

	saltBytes := [32]byte{}
	salt.FillBytes(saltBytes[:])

	saltKeccak := crypto.Keccak256(saltBytes[:], nodeAddress[:])

	nonceBytes := [32]byte{}
	nonce := m.GetSuperNodeNonce(nodeAddress)
	nonce.FillBytes(nonceBytes[:])

	sigTypeBytes := [32]byte{} // Always 0 for the mock

	message := crypto.Keccak256(
		minipoolAddress[:],
		saltKeccak[:],
		superNodeAddress[:],
		nonceBytes[:],
		sigTypeBytes[:],
		chainIdBytes[:],
	)

	// Sign the message
	signature, err := utils.CreateSignature(message, m.database.ConstellationAdminPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating signature: %w", err)
	}
	return signature, nil
}

// Get the whitelist signature nonce for a node address
func (m *NodeSetMockManager) GetWhitelistNonce(nodeAddress ethcommon.Address) *big.Int {
	nonce, exists := m.database.ConstellationWhitelistNonces[nodeAddress]
	if !exists {
		m.database.ConstellationWhitelistNonces[nodeAddress] = 0
	}
	return new(big.Int).SetUint64(nonce)
}

// Get the supernode signature nonce for a node address
func (m *NodeSetMockManager) GetSuperNodeNonce(nodeAddress ethcommon.Address) *big.Int {
	nonce, exists := m.database.ConstellationSuperNodeNonces[nodeAddress]
	if !exists {
		m.database.ConstellationSuperNodeNonces[nodeAddress] = 0
	}
	return new(big.Int).SetUint64(nonce)
}

// Increment the whitelist nonce for a node address
func (m *NodeSetMockManager) IncrementWhitelistNonce(nodeAddress ethcommon.Address) {
	m.database.ConstellationWhitelistNonces[nodeAddress]++
}

// Increment the supernode nonce for a node address
func (m *NodeSetMockManager) IncrementSuperNodeNonce(nodeAddress ethcommon.Address) {
	m.database.ConstellationSuperNodeNonces[nodeAddress]++
}
