package db

import (
	"bytes"
	"fmt"
	"io"

	"filippo.io/age"
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
	nsutils "github.com/rocket-pool/node-manager-core/utils"
)

const (
	DefaultMaxValidatorsPerUser uint = 1
)

// Info for StakeWise vaults
type StakeWiseVault struct {
	// The vault's human-readable name
	Name string

	// The vault address
	Address ethcommon.Address

	// The map of pubkeys that have been uploaded to StakeWise
	UploadedData map[beacon.ValidatorPubkey]bool

	// Index of the latest deposit data set uploaded to StakeWise
	LatestDepositDataSetIndex int

	// Latest deposit data set uploaded to StakeWise
	LatestDepositDataSet []beacon.ExtendedDepositData

	// Map of nodes to Validators for StakeWise vaults
	Validators map[ethcommon.Address]map[beacon.ValidatorPubkey]*StakeWiseValidatorInfo

	// The max number of validators per user
	MaxValidatorsPerUser uint

	deployment *StakeWiseDeployment
	db         *Database
}

// Create a new StakeWise vault
func newStakeWiseVault(deployment *StakeWiseDeployment, name string, address ethcommon.Address) *StakeWiseVault {
	return &StakeWiseVault{
		Name:                      name,
		Address:                   address,
		UploadedData:              map[beacon.ValidatorPubkey]bool{},
		LatestDepositDataSet:      []beacon.ExtendedDepositData{},
		LatestDepositDataSetIndex: 0,
		Validators:                map[ethcommon.Address]map[beacon.ValidatorPubkey]*StakeWiseValidatorInfo{},
		MaxValidatorsPerUser:      DefaultMaxValidatorsPerUser,
		deployment:                deployment,
		db:                        deployment.db,
	}
}

// Clone the StakeWise vault
func (v *StakeWiseVault) clone(deploymentClone *StakeWiseDeployment) *StakeWiseVault {
	clone := newStakeWiseVault(deploymentClone, v.Name, v.Address)
	clone.MaxValidatorsPerUser = v.MaxValidatorsPerUser
	clone.LatestDepositDataSetIndex = v.LatestDepositDataSetIndex
	clone.LatestDepositDataSet = make([]beacon.ExtendedDepositData, len(v.LatestDepositDataSet))
	copy(clone.LatestDepositDataSet, v.LatestDepositDataSet)
	for node, validators := range v.Validators {
		cloneValidators := map[beacon.ValidatorPubkey]*StakeWiseValidatorInfo{}
		for pubkey, validator := range validators {
			cloneValidators[pubkey] = validator.clone()
		}
		clone.Validators[node] = cloneValidators
	}
	for pubkey, uploaded := range v.UploadedData {
		clone.UploadedData[pubkey] = uploaded
	}
	return clone
}

// Add a new StakeWise validator to the node
func (v *StakeWiseVault) AddStakeWiseDepositData(node *Node, depositData beacon.ExtendedDepositData) {
	pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
	nodeValidators, nodeExists := v.Validators[node.Address]
	if !nodeExists {
		nodeValidators = map[beacon.ValidatorPubkey]*StakeWiseValidatorInfo{}
		v.Validators[node.Address] = nodeValidators
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
	return v.Validators[node.Address]
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
		validators := v.Validators[node.Address]
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

// Handle a new collection of encrypted signed exits from a node for StakeWise
func (v *StakeWiseVault) HandleEncryptedSignedExitUpload(node *Node, data []common.EncryptedExitData) error {
	// Add the signed exits
	for _, signedExit := range data {
		pubkey, err := beacon.HexToValidatorPubkey(signedExit.Pubkey)
		if err != nil {
			return fmt.Errorf("error parsing validator pubkey [%s]: %w", signedExit.Pubkey, err)
		}

		// Decrypt the exit data
		if v.db.secretEncryptionIdentity == nil {
			return fmt.Errorf("secret encryption identity not set")
		}
		decodedHex, err := nsutils.DecodeHex(signedExit.ExitMessage)
		if err != nil {
			return fmt.Errorf("error decoding exit message hex: %w", err)
		}
		encReader := bytes.NewReader(decodedHex)
		decReader, err := age.Decrypt(encReader, v.db.secretEncryptionIdentity)
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
		validators := v.Validators[node.Address]
		if len(validators) == 0 {
			return fmt.Errorf("vault [%s] is not used by node [%s]", v.Address.Hex(), node.Address.Hex())
		}
		found := false
		for _, validator := range validators {
			if validator.Pubkey == pubkey {
				validator.SetExitMessage(exitMessage)
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
			validators := v.Validators[node.Address]
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
				validators := v.Validators[node.Address]
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
				validators := v.Validators[node.Address]
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

// Get the number of active / registered validators for a user
func (v *StakeWiseVault) GetRegisteredValidatorsPerUser(user *User) uint {
	registered := uint(0)
	for _, node := range user.nodes {
		if !node.isRegistered {
			continue
		}
		validators := v.Validators[node.Address]
		for _, validator := range validators {
			if validator.IsActiveOnBeacon ||
				validator.HasDepositEvent ||
				validator.BeaconDepositRoot == v.db.Eth.depositRoot {
				registered++
			}
		}
	}
	return registered
}
