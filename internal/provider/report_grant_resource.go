package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                   = (*reportGrantResource)(nil)
	_ resource.ResourceWithConfigure      = (*reportGrantResource)(nil)
	_ resource.ResourceWithImportState    = (*reportGrantResource)(nil)
	_ resource.ResourceWithValidateConfig = (*reportGrantResource)(nil)
)

type reportGrantResource struct{ configuredResource }

func NewReportGrantResource() resource.Resource { return &reportGrantResource{} }

func (r *reportGrantResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_report_grant"
}

func (r *reportGrantResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.configure(request, response)
}

func (r *reportGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = grantSchema(
		"Manages a Sigma report grant.",
		"Report ID.",
		"Report permission: `view` or `edit`.",
	)
}

func (r *reportGrantResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config grantModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}
	if config.MemberID.IsUnknown() || config.TeamID.IsUnknown() || config.Permission.IsUnknown() || config.TagID.IsUnknown() {
		return
	}
	validateGrant(&config, reportGrantPermissions, true, &response.Diagnostics)
}

func (r *reportGrantResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	createDocumentGrant(ctx, request, response, r.client, "reports", reportGrantPermissions)
}

func (r *reportGrantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	readDocumentGrant(ctx, request, response, r.client)
}

func (r *reportGrantResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}

func (r *reportGrantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	deleteDocumentGrant(ctx, request, response, r.client, "reports")
}

func (r *reportGrantResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importGrantCompositeID(ctx, request, response)
}

func createDocumentGrant(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse, client *sigma.Client, kind string, allowed []string) {
	var plan grantModel
	response.Diagnostics.Append(request.Plan.Get(ctx, &plan)...)
	if response.Diagnostics.HasError() || !validateGrant(&plan, allowed, true, &response.Diagnostics) {
		return
	}
	grantee := sigma.Grantee{MemberID: plan.MemberID.ValueString(), TeamID: plan.TeamID.ValueString()}
	tagID := plan.TagID.ValueString()
	var value *sigma.Grant
	var err error
	// Prefer the generic grants API when a version tag is set so we receive the
	// created grant ID. List responses do not reliably include tagId, so
	// post-create lookups by grantee+permission alone are ambiguous.
	if tagID != "" {
		value, err = client.CreateGrant(ctx, sigma.CreateGrantInput{
			Grantee: grantee, Permission: plan.Permission.ValueString(), InodeID: plan.InodeID.ValueString(), TagID: tagID,
		})
	} else {
		err = client.CreateDocumentGrant(ctx, kind, plan.InodeID.ValueString(), grantee, plan.Permission.ValueString(), tagID)
	}
	if err != nil {
		response.Diagnostics.AddError("Unable to create Sigma grant", err.Error())
		return
	}
	if value == nil {
		value, err = lookupListedGrant(func() ([]sigma.Grant, error) {
			return client.ListGrants(ctx, plan.InodeID.ValueString())
		}, &plan, "")
		if err != nil {
			response.Diagnostics.AddError("Unable to locate created Sigma grant", err.Error())
			return
		}
	}
	setGrant(&plan, value)
	response.Diagnostics.Append(response.State.Set(ctx, &plan)...)
}

func readDocumentGrant(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse, client *sigma.Client) {
	var state grantModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	value, err := lookupListedGrant(func() ([]sigma.Grant, error) {
		return client.ListGrants(ctx, state.InodeID.ValueString())
	}, &state, state.ID.ValueString())
	if sigma.IsNotFound(err) {
		response.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		response.Diagnostics.AddError("Unable to read Sigma grant", err.Error())
		return
	}
	setGrant(&state, value)
	response.Diagnostics.Append(response.State.Set(ctx, &state)...)
}

func deleteDocumentGrant(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse, client *sigma.Client, kind string) {
	var state grantModel
	response.Diagnostics.Append(request.State.Get(ctx, &state)...)
	if response.Diagnostics.HasError() {
		return
	}
	if err := client.DeleteDocumentGrant(ctx, kind, state.InodeID.ValueString(), state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		response.Diagnostics.AddError("Unable to delete Sigma grant", err.Error())
	}
}
