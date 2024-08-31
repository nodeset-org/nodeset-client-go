package db

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/rocket-pool/node-manager-core/beacon"
)

var (
	ErrAlreadyRegistered error = errors.New("node has already been registered with the NodeSet server")
)

// A node
type Node struct {
	Address common.Address

	stakeWiseValidators map[string]map[common.Address][]*StakeWiseValidatorInfo
	isRegistered        bool
	user                *User
}

// Create a new node
func newNode(user *User, address common.Address) *Node {
	return &Node{
		Address:             address,
		stakeWiseValidators: map[string]map[common.Address][]*StakeWiseValidatorInfo{},
		user:                user,
	}
}

// clone the node
func (n *Node) clone(userClone *User) *Node {
	clone := newNode(userClone, n.Address)
	clone.isRegistered = n.isRegistered

	for deploymentID, deployment := range n.stakeWiseValidators {
		for vaultAddress, validators := range deployment {
			cloneSlice := make([]*StakeWiseValidatorInfo, len(validators))
			for i, validator := range validators {
				cloneSlice[i] = validator.clone()
			}
			cloneDeploymentMap := clone.stakeWiseValidators[deploymentID]
			if cloneDeploymentMap == nil {
				cloneDeploymentMap = map[common.Address][]*StakeWiseValidatorInfo{}
				clone.stakeWiseValidators[deploymentID] = cloneDeploymentMap
			}
			cloneDeploymentMap[vaultAddress] = cloneSlice
		}
	}
	return clone
}

// Check if the node is registered or not
func (n *Node) IsRegistered() bool {
	return n.isRegistered
}

// Register the node with the NodeSet server
func (n *Node) Register(signature []byte) error {
	return n.registerImpl(signature, false)
}

// Register the node with the NodeSet server, bypassing the signature requirement for testing
func (n *Node) RegisterWithoutSignature() error {
	return n.registerImpl(nil, true)
}

// Add a new StakeWise validator to the node
func (n *Node) AddStakeWiseDepositData(vault *StakeWiseVault, depositData beacon.ExtendedDepositData) {
	validatorsForDeployment, exists := n.stakeWiseValidators[vault.deployment.ID]
	if !exists {
		validatorsForDeployment = map[common.Address][]*StakeWiseValidatorInfo{}
		n.stakeWiseValidators[vault.deployment.ID] = validatorsForDeployment
	}

	validatorsForVault, exists := validatorsForDeployment[vault.Address]
	if !exists {
		validatorsForVault = []*StakeWiseValidatorInfo{}
		validatorsForDeployment[vault.Address] = validatorsForVault
	}

	pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
	for _, validator := range validatorsForVault {
		if validator.Pubkey == pubkey {
			// Already present
			return
		}
	}

	validator := newStakeWiseValidatorInfo(depositData, vault)
	validatorsForVault = append(validatorsForVault, validator)
	n.stakeWiseValidators[vault.deployment.ID][vault.Address] = validatorsForVault
}

// Get the StakeWise validators for the node
func (n *Node) GetStakeWiseValidatorsForVault(vault *StakeWiseVault) []*StakeWiseValidatorInfo {
	validatorsForDeployment := n.stakeWiseValidators[vault.deployment.ID]
	if validatorsForDeployment == nil {
		return nil
	}
	return validatorsForDeployment[vault.Address]
}

// Get all StakeWise validators
func (n *Node) GetAllStakeWiseValidators(deployment *StakeWiseDeployment) map[common.Address][]*StakeWiseValidatorInfo {
	return n.stakeWiseValidators[deployment.ID]
}

// Implementation for registering the node
func (n *Node) registerImpl(signature []byte, skipVerification bool) error {
	if n.isRegistered {
		return ErrAlreadyRegistered
	}

	// Verify the signature
	if !skipVerification {
		err := auth.VerifyRegistrationSignature(n.user.Email, n.Address, signature)
		if err != nil {
			return err
		}
	}

	n.isRegistered = true
	return nil
}
