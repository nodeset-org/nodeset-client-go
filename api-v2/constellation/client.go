package v2constellation

import "github.com/nodeset-org/nodeset-client-go/common"

const (
	ConstellationPrefix string = "modules/constellation/"
)

type V2ConstellationClient struct {
	commonClient *common.CommonNodeSetClient
}

func NewV2ConstellationClient(commonClient *common.CommonNodeSetClient) *V2ConstellationClient {
	return &V2ConstellationClient{
		commonClient: commonClient,
	}
}
