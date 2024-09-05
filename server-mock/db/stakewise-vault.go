package db

import (
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Info for StakeWise vaults
type StakeWiseVault struct {
	// The vault address
	Address ethcommon.Address

	// The map of pubkeys that have been uploaded to StakeWise
	UploadedData map[beacon.ValidatorPubkey]bool

	// Index of the latest deposit data set uploaded to StakeWise
	LatestDepositDataSetIndex int

	// Latest deposit data set uploaded to StakeWise
	LatestDepositDataSet []beacon.ExtendedDepositData

	// Map of nodes to validators for StakeWise vaults
	validators map[ethcommon.Address]map[beacon.ValidatorPubkey]*StakeWiseValidatorInfo

	deployment *StakeWiseDeployment
	db         *Database
}

// Create a new StakeWise vault
func newStakeWiseVault(deployment *StakeWiseDeployment, address ethcommon.Address) *StakeWiseVault {
	return &StakeWiseVault{
		Address:                   address,
		UploadedData:              map[beacon.ValidatorPubkey]bool{},
		LatestDepositDataSet:      []beacon.ExtendedDepositData{},
		LatestDepositDataSetIndex: 0,
		validators:                map[ethcommon.Address]map[beacon.ValidatorPubkey]*StakeWiseValidatorInfo{},
		deployment:                deployment,
		db:                        deployment.db,
	}
}

// Clone the StakeWise vault
func (v *StakeWiseVault) clone(deploymentClone *StakeWiseDeployment) *StakeWiseVault {
	clone := newStakeWiseVault(deploymentClone, v.Address)
	clone.LatestDepositDataSetIndex = v.LatestDepositDataSetIndex
	clone.LatestDepositDataSet = make([]beacon.ExtendedDepositData, len(v.LatestDepositDataSet))
	copy(clone.LatestDepositDataSet, v.LatestDepositDataSet)
	for node, validators := range v.validators {
		cloneValidators := map[beacon.ValidatorPubkey]*StakeWiseValidatorInfo{}
		for pubkey, validator := range validators {
			cloneValidators[pubkey] = validator.clone()
		}
		clone.validators[node] = cloneValidators
	}
	for pubkey, uploaded := range v.UploadedData {
		clone.UploadedData[pubkey] = uploaded
	}
	return clone
}

// Add a new StakeWise validator to the node
func (v *StakeWiseVault) AddStakeWiseDepositData(node *Node, depositData beacon.ExtendedDepositData) {
	pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
	nodeValidators, nodeExists := v.validators[node.Address]
	if !nodeExists {
		nodeValidators = map[beacon.ValidatorPubkey]*StakeWiseValidatorInfo{}
		v.validators[node.Address] = nodeValidators
	}
	_, exists := nodeValidators[pubkey]
	if exists {
		// Already present
		return
	}

	validator := newStakeWiseValidatorInfo(depositData)
	nodeValidators[pubkey] = validator
}

// Get the StakeWise validators for a node
func (v *StakeWiseVault) GetStakeWiseValidatorsForNode(node *Node) map[beacon.ValidatorPubkey]*StakeWiseValidatorInfo {
	return v.validators[node.Address]
}

// Handle a new collection of deposit data uploads from a node
func (v *StakeWiseVault) HandleDepositDataUpload(node *Node, data []beacon.ExtendedDepositData) error {
	// Add the deposit data
	for _, depositData := range data {
		wcAddress := ethcommon.BytesToAddress(depositData.WithdrawalCredentials)
		if wcAddress != v.Address {
			return fmt.Errorf("deposit data withdrawal credentials [%s] don't match vault address [%s]", wcAddress.Hex(), v.Address.Hex())
		}
		v.AddStakeWiseDepositData(node, depositData)
	}

	return nil
}

// Handle a new collection of signed exits from a node for StakeWise
func (v *StakeWiseVault) HandleSignedExitUpload(node *Node, data []common.ExitData) error {
	// Add the signed exits
	for _, signedExit := range data {
		pubkey, err := beacon.HexToValidatorPubkey(signedExit.Pubkey)
		if err != nil {
			return fmt.Errorf("error parsing validator pubkey [%s]: %w", signedExit.Pubkey, err)
		}

		// Get the validator
		validators := v.validators[node.Address]
		if len(validators) == 0 {
			return fmt.Errorf("vault [%s] is not used by node [%s]", v.Address.Hex(), node.Address.Hex())
		}
		found := false
		for _, validator := range validators {
			if validator.Pubkey == pubkey {
				validator.SetExitMessage(signedExit.ExitMessage)
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("node [%s] doesn't have validator [%s]", node.Address.Hex(), pubkey.Hex())
		}

	}
	return nil
}

// Create a new deposit data set
func (v *StakeWiseVault) CreateNewDepositDataSet(validatorsPerUser int) []beacon.ExtendedDepositData {
	// Iterate the users
	depositData := []beacon.ExtendedDepositData{}
	for _, user := range v.db.Core.users {
		userCount := 0
		for _, node := range user.nodes {
			if !node.isRegistered {
				continue
			}
			validators := v.validators[node.Address]
			if len(validators) == 0 {
				continue
			}
			for _, validator := range validators {
				// Add this deposit data if it hasn't been used
				if !validator.DepositDataUsed {
					depositData = append(depositData, validator.DepositData)
					userCount++
					if userCount >= validatorsPerUser {
						break
					}
				}
			}
			if userCount >= validatorsPerUser {
				break
			}
		}
	}

	return depositData
}

// Mark the deposit data for the provided validator as uploaded to StakeWise
func (v *StakeWiseVault) MarkDepositDataUploaded(pubkey beacon.ValidatorPubkey) {
	v.UploadedData[pubkey] = true
}

// Call this to "upload" a deposit data set to StakeWise
func (v *StakeWiseVault) UploadDepositDataToStakeWise(data []beacon.ExtendedDepositData) {
	for _, depositData := range data {
		pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
		v.MarkDepositDataUploaded(pubkey)
	}
}

// Call this once a deposit data set has been "uploaded" to StakeWise
func (v *StakeWiseVault) MarkDepositDataSetUploaded(data []beacon.ExtendedDepositData) {
	// Flag each deposit data as uploaded
	for _, depositData := range data {
		for _, user := range v.db.Core.users {
			for _, node := range user.nodes {
				if !node.isRegistered {
					continue
				}
				validators := v.validators[node.Address]
				if len(validators) == 0 {
					continue
				}
				for _, validator := range validators {
					if validator.Pubkey == beacon.ValidatorPubkey(depositData.PublicKey) {
						validator.DepositData = depositData
						validator.UseDepositData()
					}
				}
			}
		}
	}

	// Increment the index
	v.LatestDepositDataSet = data
	v.LatestDepositDataSetIndex++
}

// Mark the validators as registered with StakeWise
func (v *StakeWiseVault) MarkValidatorsRegistered(data []beacon.ExtendedDepositData) {
	// Flag each validator as registered
	for _, depositData := range data {
		for _, user := range v.db.Core.users {
			for _, node := range user.nodes {
				if !node.isRegistered {
					continue
				}
				validators := v.validators[node.Address]
				if len(validators) == 0 {
					continue
				}
				for _, validator := range validators {
					if validator.Pubkey == beacon.ValidatorPubkey(depositData.PublicKey) {
						validator.MarkActive()
					}
				}
			}
		}
	}
}
