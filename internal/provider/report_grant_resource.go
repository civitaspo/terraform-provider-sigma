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
		documentGrantTagMarkdown,
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
	inodeID, diags := knownString(plan.InodeID, "inode_id")
	response.Diagnostics.Append(diags...)
	permission, diags := knownString(plan.Permission, "permission")
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	grantee := grantGrantee(&plan)
	var value *sigma.Grant
	var err error
	if taggedDocumentGrant(&plan) {
		tagID, tagDiags := knownString(plan.TagID, "tag_id")
		response.Diagnostics.Append(tagDiags...)
		if response.Diagnostics.HasError() {
			return
		}
		value, err = client.CreateGrant(ctx, sigma.CreateGrantInput{
			Grantee: grantee, Permission: permission, InodeID: inodeID, TagID: tagID,
		})
		if err != nil {
			response.Diagnostics.AddError("Unable to create Sigma grant", err.Error())
			return
		}
		if value == nil || value.GrantID == "" {
			response.Diagnostics.AddError("Unable to create Sigma grant", "Generic grant create returned an empty grant ID.")
			return
		}
	} else {
		err = client.CreateDocumentGrant(ctx, kind, inodeID, grantee, permission, "")
		if err != nil {
			response.Diagnostics.AddError("Unable to create Sigma grant", err.Error())
			return
		}
		value, err = lookupListedGrant(func() ([]sigma.Grant, error) {
			return client.ListGrants(ctx, inodeID)
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
	grantID, diags := knownString(state.ID, "id")
	response.Diagnostics.Append(diags...)
	inodeID, diags := knownString(state.InodeID, "inode_id")
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	var value *sigma.Grant
	var err error
	if taggedDocumentGrant(&state) {
		value, err = client.GetGrant(ctx, grantID)
	} else {
		value, err = lookupListedGrant(func() ([]sigma.Grant, error) {
			return client.ListGrants(ctx, inodeID)
		}, &state, grantID)
	}
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
	grantID, diags := knownString(state.ID, "id")
	response.Diagnostics.Append(diags...)
	inodeID, diags := knownString(state.InodeID, "inode_id")
	response.Diagnostics.Append(diags...)
	if response.Diagnostics.HasError() {
		return
	}
	var err error
	if taggedDocumentGrant(&state) {
		err = client.DeleteGrant(ctx, grantID)
	} else {
		err = client.DeleteDocumentGrant(ctx, kind, inodeID, grantID)
	}
	if err != nil && !sigma.IsNotFound(err) {
		response.Diagnostics.AddError("Unable to delete Sigma grant", err.Error())
	}
}
