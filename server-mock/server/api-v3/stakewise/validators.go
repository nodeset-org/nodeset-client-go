package v3server_stakewise

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/rocket-pool/node-manager-core/beacon"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	v3stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
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

	numToRegister := len(body.Validators)
	available := int(deployment.MaxValidators) - int(deployment.ActiveValidators)
	if numToRegister > available {
		servermockcommon.HandleServerError(w, s.logger, fmt.Errorf("not enough available slots: requested %d, available %d", numToRegister, available))
		return
	}
	startIndex := deployment.ActiveValidators
	deployment.ActiveValidators += uint(numToRegister)

	// Must add validator to struct + exit message
	for _, validator := range body.Validators {
		pubkey := beacon.ValidatorPubkey(validator.DepositData.PublicKey)

		// Add the validator if not already present
		vault.AddStakeWiseDepositData(node, validator.DepositData)

		// Get validator reference
		nodeValidators := vault.GetStakeWiseValidatorsForNode(node)
		if vInfo, exists := nodeValidators[pubkey]; exists {
			vInfo.SetExitMessage(common.ExitMessage{
				Message:   common.ExitMessageDetails{},
				Signature: string(validator.DepositData.Signature),
			})
			vInfo.MarkActive()
		}
	}

	// TODO: Confirm with JC
	// NICE TO HAVE: https://github.com/stakewise/v3-core/blob/main/contracts/validators/ValidatorsChecker.sol#L187
	typeHash := crypto.Keccak256Hash([]byte("StakeWiseValidatorRegistration(uint256 chainId,address vault,uint256 index,uint256 count,bytes32 depositRoot)"))

	chainIDBig := deployment.ChainID
	indexBig := big.NewInt(int64(startIndex))
	countBig := big.NewInt(int64(numToRegister))

	encoded, err := abi.Arguments{
		{Type: mustType(abi.NewType("uint256", "", nil))},
		{Type: mustType(abi.NewType("address", "", nil))},
		{Type: mustType(abi.NewType("uint256", "", nil))},
		{Type: mustType(abi.NewType("uint256", "", nil))},
		{Type: mustType(abi.NewType("bytes32", "", nil))},
	}.Pack(chainIDBig, vaultAddress, indexBig, countBig, body.BeaconDepositRoot)
	if err != nil {
		servermockcommon.HandleServerError(w, s.logger, fmt.Errorf("failed to encode args: %w", err))
		return
	}

	hash := crypto.Keccak256Hash(append(typeHash.Bytes(), encoded...))

	resp := v3stakewise.PostValidatorData{
		Signature: hash.Hex(), //solidity code for stakewise
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
