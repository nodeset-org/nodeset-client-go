package db

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-client-go/utils"
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

	// Maps for signature nonces - TEMP until proper reading from an EL
	whitelistNonces map[ethcommon.Address]uint64
	superNodeNonces map[ethcommon.Address]uint64
}

// Create a new Constellation deployment
func newConstellationDeployment(id string, chainID *big.Int, whitelistAddress ethcommon.Address, superNodeAddress ethcommon.Address) *ConstellationDeployment {
	return &ConstellationDeployment{
		ID:               id,
		ChainID:          chainID,
		WhitelistAddress: whitelistAddress,
		SuperNodeAddress: superNodeAddress,
		whitelistNonces:  map[ethcommon.Address]uint64{},
		superNodeNonces:  map[ethcommon.Address]uint64{},
	}
}

// Clone the deployment
func (d *ConstellationDeployment) Clone() *ConstellationDeployment {
	clone := newConstellationDeployment(d.ID, d.ChainID, d.WhitelistAddress, d.SuperNodeAddress)
	for address, nonce := range d.whitelistNonces {
		clone.whitelistNonces[address] = nonce
	}
	for address, nonce := range d.superNodeNonces {
		clone.superNodeNonces[address] = nonce
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

// Call this to get a signature for adding the node to the Constellation whitelist
func (d *ConstellationDeployment) GetConstellationWhitelistSignature(nodeAddress ethcommon.Address) ([]byte, error) {
	if d.adminPrivateKey == nil {
		return nil, fmt.Errorf("constellation admin private key not set")
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
	return signature, nil
}

// Call this to get a signature for depositing a new minipool with Constellation
func (d *ConstellationDeployment) GetConstellationDepositSignature(nodeAddress ethcommon.Address, minipoolAddress ethcommon.Address, salt *big.Int) ([]byte, error) {
	if d.adminPrivateKey == nil {
		return nil, fmt.Errorf("constellation admin private key not set")
	}

	chainIdBytes := [32]byte{}
	d.ChainID.FillBytes(chainIdBytes[:])

	saltBytes := [32]byte{}
	salt.FillBytes(saltBytes[:])

	saltKeccak := crypto.Keccak256(saltBytes[:], nodeAddress[:])

	nonceBytes := [32]byte{}
	nonce := big.NewInt(int64(d.superNodeNonces[nodeAddress]))
	nonce.FillBytes(nonceBytes[:])

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
	return signature, nil
}
