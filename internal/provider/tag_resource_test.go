package provider_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestTagResource(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	tag := map[string]any{
		"versionTagId": "tag-1", "name": "prod", "color": "cyan", "description": "Production",
		"ownerId": "owner-1", "createdBy": "owner-1", "updatedBy": "owner-1",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z", "isArchived": false,
	}
	mock.Mux.HandleFunc("/v2/tags", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPost:
			_ = json.NewEncoder(response).Encode(map[string]any{"versionTagId": "tag-1", "name": "prod"})
		case http.MethodGet:
			_ = json.NewEncoder(response).Encode(map[string]any{"entries": []any{tag}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/tags/tag-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		response.Header().Set("Content-Type", "application/json")
		switch request.Method {
		case http.MethodPatch:
			_ = json.NewEncoder(response).Encode(tag)
		case http.MethodDelete:
			_ = json.NewEncoder(response).Encode(map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	config := documentProviderConfig(mock) + `
resource "sigma_tag" "test" {
  name        = "prod"
  color       = "cyan"
  description = "Production"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check: resource.ComposeAggregateTestCheckFunc(
			resource.TestCheckResourceAttr("sigma_tag.test", "id", "tag-1"),
			resource.TestCheckResourceAttr("sigma_tag.test", "color", "cyan"),
		),
	}}))
}

func TestAccTagResource(t *testing.T) { requireAcceptance(t) }
