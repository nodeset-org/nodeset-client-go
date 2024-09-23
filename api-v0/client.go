package apiv0

import (
	"log/slog"
	"time"

	"github.com/nodeset-org/nodeset-client-go/common"
)

// Client for interacting with the NodeSet server
type NodeSetClient struct {
	*common.CommonNodeSetClient
}

// Creates a new NodeSet client
// baseUrl: The base URL to use for the client, for example [https://nodeset.io/api]
func NewNodeSetClient(logger *slog.Logger, baseUrl string, timeout time.Duration) *NodeSetClient {
	return &NodeSetClient{
		CommonNodeSetClient: common.NewCommonNodeSetClient(logger, baseUrl, timeout),
	}
}
