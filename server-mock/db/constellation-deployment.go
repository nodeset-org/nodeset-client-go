package db

import (
	"bytes"
	"crypto/ecdsa"
	"fmt"
	"io"
	"math/big"

	"filippo.io/age"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/goccy/go-json"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/utils"
	"github.com/rocket-pool/node-manager-core/beacon"
	nsutils "github.com/rocket-pool/node-manager-core/utils"
)

var (
	// Signed exit was already uploaded
	ErrSignedExitAlreadyUploaded error = fmt.Errorf("exit already uploaded")
)

// Deployment for Constellation info
type ConstellationDeployment struct {
	// The deployment's name / ID
	ID string

	// The Ethereum chain ID this deployment is for
	ChainID *big.Int

	// Address of the Whitelist contract
	WhitelistAddress ethcommon.Address

	// Address of the SuperNodeAccount contract
	SuperNodeAddress ethcommon.Address

	// Private key for the ADMIN_ROLE account
	adminPrivateKey *ecdsa.PrivateKey

	// Map of the whitelisted nodes for each user account
	whitelistedNodeMap map[string]ethcommon.Address

	// Map of nodes to minipools
	minipools map[ethcommon.Address][]ethcommon.Address

	// Map of minipools to validators - TEMP until proper reading from an EL
	validators map[ethcommon.Address]*ConstellationValidatorInfo

	// Maps for signature nonces - TEMP until proper reading from an EL
	whitelistNonces map[ethcommon.Address]uint64
	superNodeNonces map[ethcommon.Address]uint64

	// Database handle
	db *Database
}

// Create a new Constellation deployment
func newConstellationDeployment(db *Database, id string, chainID *big.Int, whitelistAddress ethcommon.Address, superNodeAddress ethcommon.Address) *ConstellationDeployment {
	return &ConstellationDeployment{
		ID:                 id,
		ChainID:            chainID,
		WhitelistAddress:   whitelistAddress,
		SuperNodeAddress:   superNodeAddress,
		whitelistedNodeMap: map[string]ethcommon.Address{},
		minipools:          map[ethcommon.Address][]ethcommon.Address{},
		validators:         map[ethcommon.Address]*ConstellationValidatorInfo{},
		whitelistNonces:    map[ethcommon.Address]uint64{},
		superNodeNonces:    map[ethcommon.Address]uint64{},
		db:                 db,
	}
}

// Clone the deployment
func (d *ConstellationDeployment) clone(dbClone *Database) *ConstellationDeployment {
	clone := newConstellationDeployment(dbClone, d.ID, d.ChainID, d.WhitelistAddress, d.SuperNodeAddress)
	for address, nonce := range d.whitelistNonces {
		clone.whitelistNonces[address] = nonce
	}
	for address, nonce := range d.superNodeNonces {
		clone.superNodeNonces[address] = nonce
	}
	for email, address := range d.whitelistedNodeMap {
		clone.whitelistedNodeMap[email] = address
	}
	for nodeAddress, minipools := range d.minipools {
		cloneMinipools := make([]ethcommon.Address, len(minipools))
		copy(cloneMinipools, minipools)
		clone.minipools[nodeAddress] = cloneMinipools
	}
	for minipoolAddress, validator := range d.validators {
		clone.validators[minipoolAddress] = validator.clone()
	}
	clone.adminPrivateKey = d.adminPrivateKey
	return clone
}

// Get the admin private key
func (d *ConstellationDeployment) GetAdminPrivateKey() *ecdsa.PrivateKey {
	return d.adminPrivateKey
}

// Set the admin private key
func (d *ConstellationDeployment) SetAdminPrivateKey(privateKey *ecdsa.PrivateKey) {
	d.adminPrivateKey = privateKey
}

// Get the whitelist nonce for the given address
func (d *ConstellationDeployment) GetWhitelistNonce(address ethcommon.Address) uint64 {
	return d.whitelistNonces[address]
}

// Increment the whitelist nonce for the given address
func (d *ConstellationDeployment) IncrementWhitelistNonce(address ethcommon.Address) {
	d.whitelistNonces[address]++
}

// Get the SuperNodeAccount nonce for the given address
func (d *ConstellationDeployment) GetSuperNodeNonce(address ethcommon.Address) uint64 {
	return d.superNodeNonces[address]
}

// Increment the SuperNodeAccount nonce for the given address
func (d *ConstellationDeployment) IncrementSuperNodeNonce(address ethcommon.Address) {
	d.superNodeNonces[address]++
}

// Get the whitelisted address for the given user
func (d *ConstellationDeployment) GetWhitelistedAddressForUser(userEmail string) *ethcommon.Address {
	address, exists := d.whitelistedNodeMap[userEmail]
	if !exists {
		return nil
	}
	return &address
}

// Call this to get a signature for adding the node to the Constellation whitelist
func (d *ConstellationDeployment) GetWhitelistSignature(nodeAddress ethcommon.Address) ([]byte, error) {
	if d.adminPrivateKey == nil {
		return nil, fmt.Errorf("constellation admin private key not set")
	}

	node, isRegistered := d.db.Core.GetNode(nodeAddress)
	if node == nil || !isRegistered {
		return nil, fmt.Errorf("node %s not registered", nodeAddress.Hex())
	}

	whitelistedAddress, exists := d.whitelistedNodeMap[node.user.Email]
	if exists && whitelistedAddress != nodeAddress {
		return nil, fmt.Errorf("node %s already whitelisted for user %s", whitelistedAddress.Hex(), node.user.Email)
	}

	chainIdBytes := [32]byte{}
	d.ChainID.FillBytes(chainIdBytes[:])

	nonceBytes := [32]byte{}
	nonce := big.NewInt(int64(d.whitelistNonces[nodeAddress]))
	nonce.FillBytes(nonceBytes[:])

	sigTypeBytes := [32]byte{} // Always 0 for the mock

	message := crypto.Keccak256(
		nodeAddress[:],
		d.WhitelistAddress[:],
		nonceBytes[:],
		sigTypeBytes[:],
		chainIdBytes[:],
	)

	// Hash of the concatenated addresses
	signature, err := utils.CreateSignature(message, d.adminPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating signature: %w", err)
	}
	d.whitelistedNodeMap[node.user.Email] = nodeAddress
	return signature, nil
}

// Call this to get a signature for depositing a new minipool with Constellation
func (d *ConstellationDeployment) GetMinipoolDepositSignature(nodeAddress ethcommon.Address, minipoolAddress ethcommon.Address, salt *big.Int) ([]byte, error) {
	if d.adminPrivateKey == nil {
		return nil, fmt.Errorf("constellation admin private key not set")
	}

	node, isRegistered := d.db.Core.GetNode(nodeAddress)
	if node == nil || !isRegistered {
		return nil, fmt.Errorf("node %s not registered", nodeAddress.Hex())
	}

	whitelistedAddress, exists := d.whitelistedNodeMap[node.user.Email]
	if !exists || whitelistedAddress != nodeAddress {
		return nil, fmt.Errorf("node %s not set as Constellation node for %s", nodeAddress.Hex(), node.user.Email)
	}

	chainIdBytes := [32]byte{}
	d.ChainID.FillBytes(chainIdBytes[:])

	saltBytes := [32]byte{}
	salt.FillBytes(saltBytes[:])

	saltKeccak := crypto.Keccak256(saltBytes[:], nodeAddress[:])

	nonceBytes := [32]byte{}
	nonce := d.superNodeNonces[nodeAddress]
	nonceBig := big.NewInt(int64(nonce))
	nonceBig.FillBytes(nonceBytes[:])

	sigTypeBytes := [32]byte{} // Always 0 for the mock

	message := crypto.Keccak256(
		minipoolAddress[:],
		saltKeccak[:],
		d.SuperNodeAddress[:],
		nonceBytes[:],
		sigTypeBytes[:],
		chainIdBytes[:],
	)

	// Sign the message
	signature, err := utils.CreateSignature(message, d.adminPrivateKey)
	if err != nil {
		return nil, fmt.Errorf("error creating signature: %w", err)
	}

	// Set the minipool
	nodeMinipools := d.minipools[nodeAddress]
	if nonce >= uint64(len(nodeMinipools)) {
		nodeMinipools = append(nodeMinipools, minipoolAddress)
	} else {
		nodeMinipools[nonce] = minipoolAddress
	}
	d.minipools[nodeAddress] = nodeMinipools

	return signature, nil
}

// Set the validator pubkey for the minipool - TEMP until reading from an EL
func (d *ConstellationDeployment) SetValidatorInfoForMinipool(minipoolAddress ethcommon.Address, pubkey beacon.ValidatorPubkey) {
	d.validators[minipoolAddress] = newConstellationValidatorInfo(pubkey)
}

// Get the validators for the node
func (d *ConstellationDeployment) GetValidatorsForNode(node *Node) []*ConstellationValidatorInfo {
	minipools := d.minipools[node.Address]
	validatorInfos := []*ConstellationValidatorInfo{}
	for _, minipool := range minipools {
		validator, exists := d.validators[minipool]
		if !exists {
			continue
		}
		validatorInfos = append(validatorInfos, validator)
	}
	return validatorInfos
}

// Get the validator for a node with the given pubkey
func (d *ConstellationDeployment) GetValidator(node *Node, pubkey beacon.ValidatorPubkey) *ConstellationValidatorInfo {
	minipools := d.minipools[node.Address]
	for _, minipool := range minipools {
		validator, exists := d.validators[minipool]
		if !exists {
			continue
		}
		if validator.Pubkey == pubkey {
			return validator
		}
	}
	return nil
}

// Handle a new collection of signed exits from a node for Constellation
func (d *ConstellationDeployment) HandleSignedExitUpload(node *Node, data []common.ExitData) error {
	// Add the signed exits
	minipools := d.minipools[node.Address]
	for _, signedExit := range data {
		pubkey, err := beacon.HexToValidatorPubkey(signedExit.Pubkey)
		if err != nil {
			return fmt.Errorf("error parsing validator pubkey [%s]: %w", signedExit.Pubkey, err)
		}

		// Get the validator
		found := false
		for _, minipoolAddress := range minipools {
			validator := d.validators[minipoolAddress]
			if validator == nil || validator.Pubkey != pubkey {
				continue
			}
			found = true
			exitMsg := validator.GetExitMessage()
			if exitMsg != nil {
				return ErrSignedExitAlreadyUploaded
			}
			validator.SetExitMessage(&signedExit.ExitMessage)
			break
		}
		if !found {
			return fmt.Errorf("node [%s] doesn't have validator [%s]", node.Address.Hex(), pubkey.Hex())
		}

	}
	return nil
}

// Handle a new collection of encrypted signed exits from a node for Constellation
func (d *ConstellationDeployment) HandleEncryptedSignedExitUpload(node *Node, data []common.EncryptedExitData) error {
	// Add the signed exits
	minipools := d.minipools[node.Address]
	for _, signedExit := range data {
		pubkey, err := beacon.HexToValidatorPubkey(signedExit.Pubkey)
		if err != nil {
			return fmt.Errorf("error parsing validator pubkey [%s]: %w", signedExit.Pubkey, err)
		}

		// Decrypt the exit data
		if d.db.secretEncryptionIdentity == nil {
			return fmt.Errorf("secret encryption identity not set yet")
		}
		decodedHex, err := nsutils.DecodeHex(signedExit.ExitMessage)
		if err != nil {
			return fmt.Errorf("error decoding exit message hex: %w", err)
		}
		encReader := bytes.NewReader(decodedHex)
		decReader, err := age.Decrypt(encReader, d.db.secretEncryptionIdentity)
		if err != nil {
			return fmt.Errorf("error decrypting exit message: %w", err)
		}
		buffer := &bytes.Buffer{}
		_, err = io.Copy(buffer, decReader)
		if err != nil {
			return fmt.Errorf("error reading decrypted exit message: %w", err)
		}

		// Parse the exit message
		var exitMessage common.ExitMessage
		err = json.Unmarshal(buffer.Bytes(), &exitMessage)
		if err != nil {
			return fmt.Errorf("error parsing decrypted exit message: %w", err)
		}

		// Get the validator
		found := false
		for _, minipoolAddress := range minipools {
			validator := d.validators[minipoolAddress]
			if validator == nil || validator.Pubkey != pubkey {
				continue
			}
			found = true
			exitMsg := validator.GetExitMessage()
			if exitMsg != nil {
				return ErrSignedExitAlreadyUploaded
			}
			validator.SetExitMessage(&exitMessage)
			break
		}
		if !found {
			return fmt.Errorf("node [%s] doesn't have validator [%s]", node.Address.Hex(), pubkey.Hex())
		}

	}
	return nil
}
