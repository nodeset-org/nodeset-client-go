package v3server_stakewise

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"filippo.io/age"
	"github.com/goccy/go-json"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/rocket-pool/node-manager-core/beacon"
	nsutils "github.com/rocket-pool/node-manager-core/utils"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	v3stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
	"github.com/nodeset-org/nodeset-client-go/common"
	servermockcommon "github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) handleValidators(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getValidators(w, r)
	case http.MethodPost:
		s.postValidators(w, r)

	default:
		servermockcommon.HandleInvalidMethod(w, s.logger)
	}
}

// POST api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) postValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body v3stakewise.Validators_PostBody

	_, pathArgs := servermockcommon.ProcessApiRequest(s, w, r, &body)
	session := servermockcommon.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := servermockcommon.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Input validation
	db := s.manager.GetDatabase()
	deploymentID := pathArgs["deployment"]
	deployment := db.StakeWise.GetDeployment(deploymentID)
	if deployment == nil {
		servermockcommon.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	vaultAddress := ethcommon.HexToAddress(pathArgs["vault"])
	vault := deployment.GetVault(vaultAddress)
	if vault == nil {
		servermockcommon.HandleInvalidVault(w, s.logger, deploymentID, vaultAddress)
		return
	}

	// Filter out validators with empty public keys
	validValidators := make([]v3stakewise.ValidatorRegistrationDetails, 0, len(body.Validators))
	for _, v := range body.Validators {
		if len(v.DepositData.PublicKey) == 48 {
			requestValidator := v3stakewise.ValidatorRegistrationDetails{
				DepositData: beacon.ExtendedDepositData(v.DepositData),
				ExitMessage: v.ExitMessage,
			}
			validValidators = append(validValidators, requestValidator)
		}
	}

	user := node.GetUser()
	active := vault.GetRegisteredValidatorsPerUser(user)
	available := vault.MaxValidatorsPerUser - active
	numToRegister := uint(len(validValidators))
	if numToRegister > available {
		servermockcommon.HandleServerError(w, s.logger, fmt.Errorf(
			"not enough available slots: requested %d, available %d",
			numToRegister, available))
		return
	}

	// Must add validator to struct + exit message
	secret := db.GetSecretEncryptionIdentity()
	for _, validator := range validValidators {
		pubkey := beacon.ValidatorPubkey(validator.DepositData.PublicKey)

		// Add the validator if not already present
		vault.AddStakeWiseDepositData(node, validator.DepositData)

		decodedHex, err := nsutils.DecodeHex(validator.ExitMessage)
		if err != nil {
			servermockcommon.HandleServerError(w, s.logger, fmt.Errorf("error decoding exit message hex: %w", err))
			return
		}
		encReader := bytes.NewReader(decodedHex)
		decReader, err := age.Decrypt(encReader, secret)
		if err != nil {
			servermockcommon.HandleServerError(w, s.logger, fmt.Errorf("error decrypting exit message: %w", err))
			return
		}
		buffer := &bytes.Buffer{}
		_, err = io.Copy(buffer, decReader)
		if err != nil {
			servermockcommon.HandleServerError(w, s.logger, fmt.Errorf("error reading decrypted exit message: %w", err))
			return
		}

		var exitMessage common.ExitMessage
		err = json.Unmarshal(buffer.Bytes(), &exitMessage)
		if err != nil {
			servermockcommon.HandleServerError(w, s.logger, fmt.Errorf("error parsing decrypted exit message: %w", err))
			return
		}
		// Get validator reference
		nodeValidators := vault.GetStakeWiseValidatorsForNode(node)
		if vInfo, exists := nodeValidators[pubkey]; exists {
			vInfo.SetExitMessage(exitMessage)
			vInfo.MarkActive()
		}
	}

	// https://github.com/stakewise/v3-core/blob/main/contracts/validators/ValidatorsChecker.sol#L187
	// 1. Compute the domain separator
	domainTypeHash := crypto.Keccak256Hash([]byte("EIP712Domain(string name,string version,uint256 chainId,address verifyingContract)"))
	nameHash := crypto.Keccak256Hash([]byte("VaultValidators"))
	versionHash := crypto.Keccak256Hash([]byte("1"))

	domainEncoded, err := abi.Arguments{
		{Type: mustType(abi.NewType("bytes32", "", nil))},
		{Type: mustType(abi.NewType("bytes32", "", nil))},
		{Type: mustType(abi.NewType("bytes32", "", nil))},
		{Type: mustType(abi.NewType("uint256", "", nil))},
		{Type: mustType(abi.NewType("address", "", nil))},
	}.Pack(
		domainTypeHash,
		nameHash,
		versionHash,
		deployment.ChainID,
		vaultAddress,
	)
	if err != nil {
		servermockcommon.HandleServerError(w, s.logger, fmt.Errorf("failed to encode domain: %w", err))
		return
	}
	domainSeparator := crypto.Keccak256Hash(domainEncoded)

	// 2. Compute keccak256(validators)
	validatorsBytes, err := json.Marshal(body.Validators)
	if err != nil {
		servermockcommon.HandleServerError(w, s.logger, fmt.Errorf("failed to marshal validators: %w", err))
		return
	}
	validatorsHash := crypto.Keccak256Hash(validatorsBytes)

	// 3. Encode and hash the struct
	_registerValidatorsTypeHash := crypto.Keccak256Hash([]byte("VaultValidators(bytes32 validatorsRegistryRoot,bytes validators)"))

	structEncoded, err := abi.Arguments{
		{Type: mustType(abi.NewType("bytes32", "", nil))},
		{Type: mustType(abi.NewType("bytes32", "", nil))},
	}.Pack(
		body.BeaconDepositRoot,
		validatorsHash,
	)
	if err != nil {
		servermockcommon.HandleServerError(w, s.logger, fmt.Errorf("failed to encode struct: %w", err))
		return
	}
	dataToHash := append(_registerValidatorsTypeHash.Bytes(), structEncoded...)
	hashStruct := crypto.Keccak256Hash(dataToHash)

	// 4. EIP-712 final digest
	finalDigestBytes := append([]byte("\x19\x01"), domainSeparator.Bytes()...)
	finalDigestBytes = append(finalDigestBytes, hashStruct.Bytes()...)
	finalDigest := crypto.Keccak256Hash(finalDigestBytes)

	resp := v3stakewise.PostValidatorData{
		Signature: finalDigest.Hex(), //solidity code for stakewise
	}
	servermockcommon.HandleSuccess(w, s.logger, resp)
}

// GET api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) getValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	_, pathArgs := servermockcommon.ProcessApiRequest(s, w, r, nil)
	session := servermockcommon.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := servermockcommon.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Input validation
	db := s.manager.GetDatabase()
	deploymentID := pathArgs["deployment"]
	deployment := db.StakeWise.GetDeployment(deploymentID)
	if deployment == nil {
		servermockcommon.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	vaultAddress := ethcommon.HexToAddress(pathArgs["vault"])
	vault := deployment.GetVault(vaultAddress)
	if vault == nil {
		servermockcommon.HandleInvalidVault(w, s.logger, deploymentID, vaultAddress)
		return
	}

	// Find the validator
	validatorStatuses := []v3stakewise.ValidatorStatus{}
	validators := vault.GetStakeWiseValidatorsForNode(node)
	for _, validator := range validators {
		validatorStatuses = append(validatorStatuses, v3stakewise.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			ExitMessageUploaded: validator.ExitMessageUploaded,
		})
	}

	// Write the response
	data := v3stakewise.ValidatorsData{
		Validators: validatorStatuses,
	}
	servermockcommon.HandleSuccess(w, s.logger, data)
}

func mustType(t abi.Type, err error) abi.Type {
	if err != nil {
		panic(err)
	}
	return t
}
