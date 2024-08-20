package v2core

import "github.com/nodeset-org/nodeset-client-go/common"

const (
	CorePrefix string = "core/"
)

type V2CoreClient struct {
	commonClient *common.CommonNodeSetClient
}

func NewV2CoreClient(commonClient *common.CommonNodeSetClient) *V2CoreClient {
	return &V2CoreClient{
		commonClient: commonClient,
	}
}
