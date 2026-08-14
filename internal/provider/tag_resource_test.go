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
		switch request.Method {
		case http.MethodPost:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			if payload["description"] != "Production" {
				t.Errorf("create description = %#v", payload["description"])
			}
			writeJSON(response, map[string]any{"versionTagId": "tag-1", "name": "prod"})
		case http.MethodGet:
			writeJSON(response, map[string]any{"entries": []any{tag}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/tags/tag-1", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPatch:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			if payload["description"] != "Updated" {
				t.Errorf("update description = %#v", payload["description"])
			}
			tag["description"] = "Updated"
			writeJSON(response, tag)
		case http.MethodDelete:
			writeJSON(response, map[string]any{})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	create := documentProviderConfig(mock) + `
resource "sigma_tag" "test" {
  name        = "prod"
  color       = "cyan"
  description = "Production"
}
`
	update := documentProviderConfig(mock) + `
resource "sigma_tag" "test" {
  name        = "prod"
  color       = "cyan"
  description = "Updated"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{
		{
			Config: create,
			Check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("sigma_tag.test", "id", "tag-1"),
				resource.TestCheckResourceAttr("sigma_tag.test", "color", "cyan"),
			),
		},
		{
			Config: update,
			Check:  resource.TestCheckResourceAttr("sigma_tag.test", "description", "Updated"),
		},
		{
			ResourceName:      "sigma_tag.test",
			ImportState:       true,
			ImportStateVerify: true,
		},
	}))
}

func TestTagResourceOmitsNullDescription(t *testing.T) {
	mock := testutil.NewMockSigma(t)
	tag := map[string]any{
		"versionTagId": "tag-1", "name": "prod", "color": "cyan",
		"ownerId": "owner-1", "createdBy": "owner-1", "updatedBy": "owner-1",
		"createdAt": "2024-01-01T00:00:00Z", "updatedAt": "2024-01-01T00:00:00Z", "isArchived": false,
	}
	mock.Mux.HandleFunc("/v2/tags", func(response http.ResponseWriter, request *http.Request) {
		mock.AssertBearer(t, request)
		switch request.Method {
		case http.MethodPost:
			var payload map[string]any
			_ = json.NewDecoder(request.Body).Decode(&payload)
			if _, ok := payload["description"]; ok {
				t.Errorf("create unexpectedly sent description: %#v", payload)
			}
			writeJSON(response, map[string]any{"versionTagId": "tag-1", "name": "prod", "color": "cyan"})
		case http.MethodGet:
			writeJSON(response, map[string]any{"entries": []any{tag}, "nextPage": nil})
		default:
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
		}
	})
	mock.Mux.HandleFunc("/v2/tags/tag-1", func(response http.ResponseWriter, request *http.Request) {
		if request.Method != http.MethodDelete {
			http.Error(response, "unexpected method", http.StatusMethodNotAllowed)
			return
		}
		writeJSON(response, map[string]any{})
	})
	config := documentProviderConfig(mock) + `
resource "sigma_tag" "test" {
  name  = "prod"
  color = "cyan"
}
`
	resource.UnitTest(t, documentTestCase([]resource.TestStep{{
		Config: config,
		Check:  resource.TestCheckResourceAttr("sigma_tag.test", "id", "tag-1"),
	}}))
}

func TestAccTagResource(t *testing.T) { runAccTag(t) }
