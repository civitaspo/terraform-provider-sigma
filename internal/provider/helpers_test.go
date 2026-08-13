package provider_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	sigmaprovider "github.com/civitaspo/terraform-provider-sigma/internal/provider"
	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func providerConfig(mock *testutil.MockSigma) string {
	return `
provider "sigma" {
  base_url      = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
}

func identityProviderConfig(mock *testutil.MockSigma) string {
	return providerConfig(mock)
}

func connectionProviderConfig(mock *testutil.MockSigma) string {
	return `
provider "sigma" {
  base_url     = "` + mock.URL() + `"
  client_id     = "` + mock.ClientID + `"
  client_secret = "` + mock.ClientSecret + `"
}
`
}

func documentProviderConfig(mock *testutil.MockSigma) string {
	return providerConfig(mock)
}

func betaProviderConfig(mock *testutil.MockSigma) string {
	return providerConfig(mock)
}

func providerTestCase(steps []resource.TestStep) resource.TestCase {
	return resource.TestCase{
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: steps,
	}
}

func identityTestCase(steps []resource.TestStep) resource.TestCase {
	return providerTestCase(steps)
}

func connectionTestCase(steps []resource.TestStep) resource.TestCase {
	return providerTestCase(steps)
}

func documentTestCase(steps []resource.TestStep) resource.TestCase {
	return providerTestCase(steps)
}

func betaTestCase(steps []resource.TestStep) resource.TestCase {
	return providerTestCase(steps)
}

func writeJSON(response http.ResponseWriter, value any) {
	response.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(response).Encode(value)
}

func nullable(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func requireAcceptance(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run Sigma acceptance tests")
	}
	t.Skip("acceptance test requires dedicated Sigma test fixtures")
}
