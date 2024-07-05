package apiv2

import (
	"net/url"
	"time"

	apiv1 "github.com/nodeset-org/nodeset-client-go/api-v1"
)

const (
	// API version to use
	ApiVersion string = "v2"

	CorePath          string = "core/"
	StakeWisePath     string = "modules/stakewise/"
	ConstellationPath string = "modules/constellation/"
)

// List of routes for v2 API functions
type V2Routes struct {
	MinipoolAvailable        string
	MinipoolDepositSignature string
	Whitelist                string
}

// Client for interacting with the NodeSet server
type NodeSetClient struct {
	*apiv1.NodeSetClient
	routes V2Routes
}

// Creates a new NodeSet client
// baseUrl: The base URL to use for the client, for example [https://nodeset.io/api]
func NewNodeSetClient(baseUrl string, timeout time.Duration) *NodeSetClient {
	expandedUrl, _ := url.JoinPath(baseUrl, ApiVersion) // becomes [https://nodeset.io/api/v2]
	client := &NodeSetClient{
		NodeSetClient: apiv1.NewNodeSetClient(expandedUrl, timeout),
		routes: V2Routes{
			MinipoolAvailable:        ConstellationPath + MinipoolAvailablePath,
			MinipoolDepositSignature: ConstellationPath + MinipoolDepositSignaturePath,
			Whitelist:                ConstellationPath + WhitelistPath,
		},
	}
	client.SetRoutes(apiv1.V1Routes{
		Login:           CorePath + apiv1.LoginPath,
		Nonce:           CorePath + apiv1.NoncePath,
		NodeAddress:     CorePath + apiv1.NodeAddressPath,
		DepositData:     StakeWisePath + apiv1.DepositDataPath,
		DepositDataMeta: StakeWisePath + apiv1.DepositDataMetaPath,
		Validators:      StakeWisePath + apiv1.ValidatorsPath,
	})
	return client
}
