package v2stakewise

import "github.com/nodeset-org/nodeset-client-go/common"

const (
	StakeWisePrefix string = "modules/stakewise/"
)

type V2StakeWiseClient struct {
	commonClient *common.CommonNodeSetClient
}

func NewV2StakeWiseClient(commonClient *common.CommonNodeSetClient) *V2StakeWiseClient {
	return &V2StakeWiseClient{
		commonClient: commonClient,
	}
}
