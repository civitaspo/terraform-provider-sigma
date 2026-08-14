package sigma_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/civitaspo/terraform-provider-sigma/internal/provider/testutil"
	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
)

func TestConnectionsClientContract(t *testing.T) {
	t.Parallel()
	conn := map[string]any{"connectionId": "conn-1", "name": "warehouse", "type": "postgres"}
	grant := map[string]any{"grantId": "grant-1", "inodeId": "conn-1", "permission": "usage"}
	path := map[string]any{"inodeId": "path-1", "name": "PUBLIC"}
	connector := map[string]any{"apiConnectorId": "ac-1", "name": "HTTP"}
	credential := map[string]any{"apiCredentialId": "cred-1", "name": "key", "authMethod": "header"}
	page := func(entry any) map[string]any { return map[string]any{"entries": []any{entry}, "nextPage": nil} }
	tokenPage := func(entry any) map[string]any { return map[string]any{"entries": []any{entry}, "nextPageToken": nil} }
	details, _ := json.Marshal(map[string]any{"host": "db.example.com"})
	mock := testutil.NewRecordingSigma(t,
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/connections", JSONBody: map[string]any{"details": map[string]any{"host": "db.example.com"}, "name": "warehouse"}, Response: conn},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/connections/conn-1", Response: conn},
		testutil.ExpectedRequest{Method: "PUT", Path: "/v2/connections/conn-1", JSONBody: map[string]any{"details": map[string]any{"host": "db.example.com"}, "name": "warehouse-2"}, Response: conn},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/connections", Response: page(conn)},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/connections/conn-1/test", Response: map[string]any{"read": "ok", "write": "ok"}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/connections/conn-1/grants", JSONBody: map[string]any{"grants": []any{map[string]any{"grantee": map[string]any{"memberId": "member-1"}, "permission": "usage"}}}, Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/connections/conn-1/grants", Response: page(grant)},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/connections/conn-1/grants/grant-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/connections/paths", Query: map[string]string{"connectionId": "conn-1"}, Response: page(path)},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/connections/paths/path-1/grants", JSONBody: map[string]any{"grants": []any{map[string]any{"grantee": map[string]any{"teamId": "team-1"}, "permission": "usage"}}}, Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/connections/paths/path-1/grants", Response: page(grant)},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/connections/paths/path-1/grants/grant-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/connection/conn-1/lookup", JSONBody: map[string]any{"path": []any{"PUBLIC", "T"}}, Response: map[string]any{"kind": "table", "inodeId": "inode-1", "url": "sigma://t"}},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/connections/conn-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/api-connectors", JSONBody: map[string]any{"name": "HTTP"}, Response: connector},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/api-connectors/ac-1", Response: connector},
		testutil.ExpectedRequest{Method: "PATCH", Path: "/v2/api-connectors/ac-1", JSONBody: map[string]any{"name": "HTTP 2"}, Response: connector},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/api-connectors", Response: tokenPage(connector)},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/api-connectors/ac-1", Response: map[string]any{}},
		testutil.ExpectedRequest{Method: "POST", Path: "/v2/api-credentials", JSONBody: map[string]any{"name": "key", "allowlist": []any{"https://example.com"}}, Response: credential},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/api-credentials/cred-1", Response: credential},
		testutil.ExpectedRequest{Method: "PATCH", Path: "/v2/api-credentials/cred-1", JSONBody: map[string]any{"name": "key-2", "allowlist": []any{"https://example.com"}}, Response: credential},
		testutil.ExpectedRequest{Method: "GET", Path: "/v2/api-credentials", Response: tokenPage(credential)},
		testutil.ExpectedRequest{Method: "DELETE", Path: "/v2/api-credentials/cred-1", Response: map[string]any{}},
	)
	client, err := sigma.NewClient(mock.URL(), mock.ClientID, mock.ClientSecret)
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if _, err := client.CreateConnection(ctx, sigma.ConnectionInput{Name: "warehouse", Details: details}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetConnection(ctx, "conn-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateConnection(ctx, "conn-1", sigma.ConnectionInput{Name: "warehouse-2", Details: details}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListConnections(ctx, sigma.ListConnectionsOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.TestConnection(ctx, "conn-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.CreateConnectionGrant(ctx, "conn-1", "member-1", "", "usage"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListConnectionGrants(ctx, "conn-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteConnectionGrant(ctx, "conn-1", "grant-1"); err != nil {
		t.Fatal(err)
	}
	cid := "conn-1"
	if _, err := client.ListConnectionPaths(ctx, sigma.ListConnectionPathsOptions{ConnectionID: &cid}); err != nil {
		t.Fatal(err)
	}
	if err := client.CreateConnectionPathGrant(ctx, "path-1", "", "team-1", "usage"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListConnectionPathGrants(ctx, "path-1"); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteConnectionPathGrant(ctx, "path-1", "grant-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.LookupConnection(ctx, "conn-1", []string{"PUBLIC", "T"}); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteConnection(ctx, "conn-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateAPIConnector(ctx, sigma.APIConnectorInput{Name: "HTTP"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetAPIConnector(ctx, "ac-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateAPIConnector(ctx, "ac-1", sigma.APIConnectorInput{Name: "HTTP 2"}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListAPIConnectors(ctx); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteAPIConnector(ctx, "ac-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.CreateAPICredential(ctx, sigma.APICredentialInput{Name: "key", Allowlist: []string{"https://example.com"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.GetAPICredential(ctx, "cred-1"); err != nil {
		t.Fatal(err)
	}
	if _, err := client.UpdateAPICredential(ctx, "cred-1", sigma.APICredentialInput{Name: "key-2", Allowlist: []string{"https://example.com"}}); err != nil {
		t.Fatal(err)
	}
	if _, err := client.ListAPICredentials(ctx); err != nil {
		t.Fatal(err)
	}
	if err := client.DeleteAPICredential(ctx, "cred-1"); err != nil {
		t.Fatal(err)
	}
}
