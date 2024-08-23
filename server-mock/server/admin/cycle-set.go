package admin

import (
	"fmt"
	"net/http"
	"strconv"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Cycle a new StakeWise deposit data set by creating it and marking it as uploaded
func (s *AdminServer) cycleSet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	deployment := query.Get("deployment")
	if deployment == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing deployment query parameter"))
		return
	}
	vaultAddressString := query.Get("vault")
	if vaultAddressString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing vault query parameter"))
		return
	}
	vaultAddress := ethcommon.HexToAddress(vaultAddressString)
	userLimit := query.Get("user-limit")
	if userLimit == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing user-limit query parameter"))
		return
	}
	validatorsPerUser, err := strconv.ParseInt(userLimit, 10, 32)
	if err != nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("error parsing user-limit: %w", err))
		return
	}

	// Create a new deposit data set
	set := s.manager.CreateNewDepositDataSet(deployment, int(validatorsPerUser))
	s.logger.Info("Created new deposit data set",
		"deployment", deployment,
		"user-limit", validatorsPerUser,
	)

	err = s.manager.UploadDepositDataToStakeWise(deployment, vaultAddress, set)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Uploaded deposit data set", "vault", vaultAddress.Hex())

	err = s.manager.MarkDepositDataSetUploaded(deployment, vaultAddress, set)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}

	vault := s.manager.GetStakeWiseVault(deployment, vaultAddress)
	if vault != nil {
		s.logger.Info("Marked deposit data set as uploaded", "version", vault.LatestDepositDataSetIndex)
	}
	common.HandleSuccess(w, s.logger, "")
}
