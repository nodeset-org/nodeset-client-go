package apiv3

import (
	"net/url"
	"time"

	v3constellation "github.com/nodeset-org/nodeset-client-go/api-v3/constellation"
	v3core "github.com/nodeset-org/nodeset-client-go/api-v3/core"
	v3stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
	"github.com/nodeset-org/nodeset-client-go/common"
)

const (
	// API version to use
	ApiVersion string = "v3"
)

// Client for interacting with the NodeSet server
type NodeSetClient struct {
	*common.CommonNodeSetClient

	// Core routes
	Core *v3core.V3CoreClient

	// StakeWise routes
	StakeWise *v3stakewise.V3StakeWiseClient

	// Constellation routes
	Constellation *v3constellation.V3ConstellationClient
}

// Creates a new NodeSet client
// baseUrl: The base URL to use for the client, for example [https://nodeset.io/api]
func NewNodeSetClient(baseUrl string, timeout time.Duration) *NodeSetClient {
	expandedUrl, _ := url.JoinPath(baseUrl, ApiVersion) // becomes [https://nodeset.io/api/v2]
	commonClient := common.NewCommonNodeSetClient(expandedUrl, timeout)
	return &NodeSetClient{
		CommonNodeSetClient: commonClient,
		Core:                v3core.NewV3CoreClient(commonClient),
		StakeWise:           v3stakewise.NewV3StakeWiseClient(commonClient),
		Constellation:       v3constellation.NewV3ConstellationClient(commonClient),
	}
}
