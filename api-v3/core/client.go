package v3core

import "github.com/nodeset-org/nodeset-client-go/common"

const (
	CorePrefix string = "core/"
)

type V3CoreClient struct {
	commonClient *common.CommonNodeSetClient
}

func NewV3CoreClient(commonClient *common.CommonNodeSetClient) *V3CoreClient {
	return &V3CoreClient{
		commonClient: commonClient,
	}
}
