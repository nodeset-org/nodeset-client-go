package auth

import (
	"crypto/ecdsa"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
	nsutil "github.com/nodeset-org/nodeset-client-go/utils"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/rocket-pool/node-manager-core/utils"
	"github.com/stretchr/testify/require"
)

// =============
// === Tests ===
// =============

func TestRecoverPubkey(t *testing.T) {
	// Get a private key
	privateKey, err := test.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}

	// Get the pubkey for it
	pubkey := crypto.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("Constructed private key, pubkey = %s", pubkey.Hex())

	// Sign a message
	message := []byte("hello world")
	signature, err := nsutil.CreateSignature(message, privateKey)
	require.NoError(t, err)
	t.Logf("Signed message, signature = %x", signature)

	// Get the pubkey from the signature
	recoveredPubkey, err := getAddressFromSignature(message, signature)
	require.NoError(t, err)
	require.Equal(t, pubkey, recoveredPubkey)
	t.Logf("Recovered pubkey matches, %s", recoveredPubkey.Hex())
}

func TestGoodRequest(t *testing.T) {
	// Get a private key
	privateKey, err := test.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}

	// Get the pubkey for it
	pubkey := crypto.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("Constructed private key, pubkey = %s", pubkey.Hex())

	// Create a request with the proper header
	vault := utils.RemovePrefix(test.StakeWiseVaultAddressHex)
	params := map[string]string{
		"vault":   vault,
		"network": test.Network,
	}
	request, expectedToken, err := generateRequest(privateKey, http.MethodGet, nil, params, "deposit-data", "meta")
	if err != nil {
		t.Fatalf("error generating request: %v", err)
	}
	t.Log("Generated deposit-data/meta request")

	// Verify the request
	token, err := GetSessionTokenFromRequest(request)
	if err != nil {
		t.Fatalf("error getting session token from request: %v", err)
	}
	require.Equal(t, expectedToken, token)
	t.Logf("Token matches (%s)", token)
}

func TestRegistration(t *testing.T) {
	// Get a private key
	privateKey, err := test.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}

	// Get the pubkey for it
	pubkey := crypto.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("Constructed private key, pubkey = %s", pubkey.Hex())

	// Sign a registration message
	email := test.User0Email
	signature, err := GetSignatureForRegistration(email, pubkey, privateKey, apiv0.NodeAddressMessageFormat)
	require.NoError(t, err)
	t.Logf("Signed registration message, signature = %x", signature)

	// Verify the signature
	err = VerifyRegistrationSignature(email, pubkey, signature, apiv0.NodeAddressMessageFormat)
	require.NoError(t, err)
	t.Log("Verified registration signature")
}

func TestLogin(t *testing.T) {
	// Get a private key
	privateKey, err := test.GetEthPrivateKey(0)
	if err != nil {
		t.Fatalf("error getting private key: %v", err)
	}

	// Get the pubkey for it
	pubkey := crypto.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("Constructed private key, pubkey = %s", pubkey.Hex())

	// Sign a login message
	nonce := "nonce"
	signature, err := GetSignatureForLogin(nonce, pubkey, privateKey)
	require.NoError(t, err)
	t.Logf("Signed login message, signature = %x", signature)

	// Verify the signature
	err = VerifyLoginSignature(nonce, pubkey, signature)
	require.NoError(t, err)
	t.Log("Verified login signature")
}

// ==========================
// === Internal Functions ===
// ==========================

// Generate an HTTP request with the signed auth header
func generateRequest(privateKey *ecdsa.PrivateKey, method string, body io.Reader, queryParams map[string]string, subroutes ...string) (*http.Request, string, error) {

	// Make the request
	path, err := url.JoinPath("http://dummy", subroutes...)
	if err != nil {
		return nil, "", fmt.Errorf("error joining path [%v]: %w", subroutes, err)
	}
	request, err := http.NewRequest(method, path, body)
	if err != nil {
		return nil, "", fmt.Errorf("error generating request to [%s]: %w", path, err)
	}
	query := request.URL.Query()
	for name, value := range queryParams {
		query.Add(name, value)
	}
	request.URL.RawQuery = query.Encode()

	// Add the auth header
	token := "token"
	AddAuthorizationHeader(request, token)
	if err != nil {
		return nil, "", fmt.Errorf("error adding auth header: %w", err)
	}
	return request, token, nil
}
