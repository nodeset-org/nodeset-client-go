package v3constellation

import "github.com/nodeset-org/nodeset-client-go/common"

const (
	ConstellationPrefix string = "modules/constellation/"
)

type V3ConstellationClient struct {
	commonClient *common.CommonNodeSetClient
}

func NewV3ConstellationClient(commonClient *common.CommonNodeSetClient) *V3ConstellationClient {
	return &V3ConstellationClient{
		commonClient: commonClient,
	}
}
