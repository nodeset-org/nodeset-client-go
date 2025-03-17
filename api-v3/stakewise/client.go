package v3stakewise

import "github.com/nodeset-org/nodeset-client-go/common"

const (
	StakeWisePrefix string = "modules/stakewise/"
)

type V3StakeWiseClient struct {
	commonClient *common.CommonNodeSetClient
}

func NewV3StakeWiseClient(commonClient *common.CommonNodeSetClient) *V3StakeWiseClient {
	return &V3StakeWiseClient{
		commonClient: commonClient,
	}
}
