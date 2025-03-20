package v3server_stakewise_test

// // Run a GET api/deposit-data request
// func runGetDepositDataRequest(t *testing.T, session *db.Session) stakewise.DepositDataData {
// 	// Create the client
// 	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
// 	client.SetSessionToken(session.Token)

// 	// Run the request
// 	data, err := client.StakeWise.DepositData_Get(context.Background(), logger, test.Network, test.StakeWiseVaultAddress)
// 	require.NoError(t, err)
// 	t.Logf("Ran request")
// 	return data
// }

// // Run a POST api/deposit-data request
// func runUploadDepositDataRequest(t *testing.T, session *db.Session, depositData []beacon.ExtendedDepositData) {
// 	// Create the client
// 	client := apiv3.NewNodeSetClient(fmt.Sprintf("http://localhost:%d/api", port), timeout)
// 	client.SetSessionToken(session.Token)

// 	// Run the request
// 	err := client.StakeWise.DepositData_Post(context.Background(), logger, test.Network, test.StakeWiseVaultAddress, depositData)
// 	require.NoError(t, err)
// 	t.Logf("Ran request")
// }
