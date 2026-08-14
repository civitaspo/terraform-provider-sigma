package provider_test

import (
	"encoding/json"
	"net/http"
	"os"
	"regexp"
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

func writeNotFound(response http.ResponseWriter) {
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(http.StatusNotFound)
	_ = json.NewEncoder(response).Encode(map[string]any{"message": "not found"})
}

func TestProviderMissingCredentials(t *testing.T) {
	t.Setenv("SIGMA_BASE_URL", "")
	t.Setenv("SIGMA_CLIENT_ID", "")
	t.Setenv("SIGMA_CLIENT_SECRET", "")
	resource.UnitTest(t, resource.TestCase{
		IsUnitTest: true,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config: `
provider "sigma" {}
data "sigma_whoami" "test" {}
`,
			ExpectError: regexp.MustCompile(`Missing`),
		}},
	})
}

func TestProviderInvalidBaseURL(t *testing.T) {
	resource.UnitTest(t, resource.TestCase{
		IsUnitTest: true,
		ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
			"sigma": providerserver.NewProtocol6WithError(sigmaprovider.New("test")()),
		},
		Steps: []resource.TestStep{{
			Config: `
provider "sigma" {
  base_url      = "://not-a-url"
  client_id     = "id"
  client_secret = "secret"
}
data "sigma_whoami" "test" {}
`,
			ExpectError: regexp.MustCompile(`(?i)invalid`),
		}},
	})
}

func requireAcceptance(t *testing.T) {
	t.Helper()
	if os.Getenv("TF_ACC") == "" {
		t.Skip("set TF_ACC=1 to run Sigma acceptance tests")
	}
	if os.Getenv("SIGMA_CLIENT_ID") == "" || os.Getenv("SIGMA_CLIENT_SECRET") == "" {
		t.Skip("SIGMA_CLIENT_ID and SIGMA_CLIENT_SECRET are required for acceptance tests")
	}
}

func requireAcceptanceEnv(t *testing.T, keys ...string) {
	t.Helper()
	requireAcceptance(t)
	for _, key := range keys {
		if os.Getenv(key) == "" {
			t.Skip(key + " is required for this acceptance test")
		}
	}
}

func assertExactQuery(t *testing.T, request *http.Request, want map[string]string, cursorKeys ...string) {
	t.Helper()
	query := request.URL.Query()
	for _, key := range cursorKeys {
		delete(query, key)
	}
	got := map[string]string{}
	for key, values := range query {
		if len(values) == 1 {
			got[key] = values[0]
		} else {
			t.Errorf("query %s = %v", key, values)
		}
	}
	if len(got) != len(want) {
		t.Errorf("query = %#v, want %#v", got, want)
	}
	for key, value := range want {
		if got[key] != value {
			t.Errorf("query[%s] = %q, want %q", key, got[key], value)
		}
	}
}

type listDataSourceCase struct {
	path       string
	cursorKey  string
	config     string
	wantQuery  map[string]string
	entry      any
	address    string
	countAttr  string
	tokenPaged bool
}

func runListDataSourceCases(t *testing.T, spec listDataSourceCase) {
	t.Helper()
	if spec.cursorKey == "" {
		spec.cursorKey = "page"
	}
	t.Run("filters", func(t *testing.T) {
		mock := testutil.NewMockSigma(t)
		mock.Mux.HandleFunc(spec.path, func(response http.ResponseWriter, request *http.Request) {
			mock.AssertBearer(t, request)
			assertExactQuery(t, request, spec.wantQuery, spec.cursorKey, "limit", "pageSize")
			writeJSON(response, map[string]any{"entries": []any{spec.entry}, "nextPage": nil, "nextPageToken": nil})
		})
		resource.UnitTest(t, providerTestCase([]resource.TestStep{{
			Config: providerConfig(mock) + spec.config,
			Check:  resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(spec.address, spec.countAttr, "1")),
		}}))
	})
	t.Run("empty", func(t *testing.T) {
		mock := testutil.NewMockSigma(t)
		mock.Mux.HandleFunc(spec.path, func(response http.ResponseWriter, request *http.Request) {
			mock.AssertBearer(t, request)
			writeJSON(response, map[string]any{"entries": []any{}, "nextPage": nil, "nextPageToken": nil})
		})
		resource.UnitTest(t, providerTestCase([]resource.TestStep{{
			Config: providerConfig(mock) + spec.config,
			Check:  resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(spec.address, spec.countAttr, "0")),
		}}))
	})
	t.Run("two_pages", func(t *testing.T) {
		mock := testutil.NewMockSigma(t)
		mock.Mux.HandleFunc(spec.path, func(response http.ResponseWriter, request *http.Request) {
			mock.AssertBearer(t, request)
			cursor := request.URL.Query().Get(spec.cursorKey)
			switch cursor {
			case "":
				if spec.tokenPaged {
					writeJSON(response, map[string]any{"entries": []any{spec.entry}, "nextPageToken": "p2"})
				} else {
					writeJSON(response, map[string]any{"entries": []any{spec.entry}, "nextPage": "p2"})
				}
			case "p2":
				writeJSON(response, map[string]any{"entries": []any{spec.entry}, "nextPage": nil, "nextPageToken": nil})
			default:
				t.Errorf("unexpected cursor %q", cursor)
				http.Error(response, "bad page", http.StatusBadRequest)
			}
		})
		resource.UnitTest(t, providerTestCase([]resource.TestStep{{
			Config: providerConfig(mock) + spec.config,
			Check:  resource.ComposeAggregateTestCheckFunc(resource.TestCheckResourceAttr(spec.address, spec.countAttr, "2")),
		}}))
	})
	t.Run("error", func(t *testing.T) {
		mock := testutil.NewMockSigma(t)
		mock.Mux.HandleFunc(spec.path, func(response http.ResponseWriter, request *http.Request) {
			mock.AssertBearer(t, request)
			response.Header().Set("Content-Type", "application/json")
			response.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(response).Encode(map[string]any{"message": "boom"})
		})
		resource.UnitTest(t, providerTestCase([]resource.TestStep{{
			Config:      providerConfig(mock) + spec.config,
			ExpectError: regexp.MustCompile("boom|Unable to list|500"),
		}}))
	})
}
