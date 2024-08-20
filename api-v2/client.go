package apiv2

import (
	"net/url"
	"time"

	v2core "github.com/nodeset-org/nodeset-client-go/api-v2/core"
	v2stakewise "github.com/nodeset-org/nodeset-client-go/api-v2/stakewise"
	"github.com/nodeset-org/nodeset-client-go/common"
)

const (
	// API version to use
	ApiVersion string = "v2"
)

// Client for interacting with the NodeSet server
type NodeSetClient struct {
	*common.CommonNodeSetClient

	// Core routes
	Core *v2core.V2CoreClient

	// StakeWise routes
	StakeWise *v2stakewise.V2StakeWiseClient
}

// Creates a new NodeSet client
// baseUrl: The base URL to use for the client, for example [https://nodeset.io/api]
func NewNodeSetClient(baseUrl string, timeout time.Duration) *NodeSetClient {
	expandedUrl, _ := url.JoinPath(baseUrl, ApiVersion) // becomes [https://nodeset.io/api/v2]
	commonClient := common.NewCommonNodeSetClient(expandedUrl, timeout)
	return &NodeSetClient{
		CommonNodeSetClient: commonClient,
		Core:                v2core.NewV2CoreClient(commonClient),
		StakeWise:           v2stakewise.NewV2StakeWiseClient(commonClient),
	}
}
