package apiv2

import (
	"net/url"
	"time"

	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
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
	*apiv0.NodeSetClient
	routes V2Routes
}

// Creates a new NodeSet client
// baseUrl: The base URL to use for the client, for example [https://nodeset.io/api]
func NewNodeSetClient(baseUrl string, timeout time.Duration) *NodeSetClient {
	expandedUrl, _ := url.JoinPath(baseUrl, ApiVersion) // becomes [https://nodeset.io/api/v2]
	client := &NodeSetClient{
		NodeSetClient: apiv0.NewNodeSetClient(expandedUrl, timeout),
		routes: V2Routes{
			MinipoolAvailable:        ConstellationPath + MinipoolAvailablePath,
			MinipoolDepositSignature: ConstellationPath + MinipoolDepositSignaturePath,
			Whitelist:                ConstellationPath + WhitelistPath,
		},
	}
	client.SetRoutes(apiv0.V1Routes{
		Login:           CorePath + apiv0.LoginPath,
		Nonce:           CorePath + apiv0.NoncePath,
		NodeAddress:     CorePath + apiv0.NodeAddressPath,
		DepositData:     StakeWisePath + apiv0.DepositDataPath,
		DepositDataMeta: StakeWisePath + apiv0.DepositDataMetaPath,
		Validators:      StakeWisePath + apiv0.ValidatorsPath,
	})
	return client
}
