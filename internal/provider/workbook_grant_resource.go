package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

var (
	_ resource.Resource                   = (*workbookGrantResource)(nil)
	_ resource.ResourceWithConfigure      = (*workbookGrantResource)(nil)
	_ resource.ResourceWithImportState    = (*workbookGrantResource)(nil)
	_ resource.ResourceWithValidateConfig = (*workbookGrantResource)(nil)
)

type workbookGrantResource struct{ configuredResource }

func NewWorkbookGrantResource() resource.Resource { return &workbookGrantResource{} }

func (r *workbookGrantResource) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_workbook_grant"
}

func (r *workbookGrantResource) Configure(_ context.Context, request resource.ConfigureRequest, response *resource.ConfigureResponse) {
	r.configure(request, response)
}

func (r *workbookGrantResource) Schema(_ context.Context, _ resource.SchemaRequest, response *resource.SchemaResponse) {
	response.Schema = grantSchema(
		"Manages a Sigma workbook grant.",
		"Workbook ID.",
		"Workbook permission: `view`, `explore`, or `edit`.",
		documentGrantTagMarkdown,
	)
}

func (r *workbookGrantResource) ValidateConfig(ctx context.Context, request resource.ValidateConfigRequest, response *resource.ValidateConfigResponse) {
	var config grantModel
	response.Diagnostics.Append(request.Config.Get(ctx, &config)...)
	if response.Diagnostics.HasError() {
		return
	}
	if config.MemberID.IsUnknown() || config.TeamID.IsUnknown() || config.Permission.IsUnknown() || config.TagID.IsUnknown() {
		return
	}
	validateGrant(&config, workbookGrantPermissions, true, &response.Diagnostics)
}

func (r *workbookGrantResource) Create(ctx context.Context, request resource.CreateRequest, response *resource.CreateResponse) {
	createDocumentGrant(ctx, request, response, r.client, "workbooks", workbookGrantPermissions)
}

func (r *workbookGrantResource) Read(ctx context.Context, request resource.ReadRequest, response *resource.ReadResponse) {
	readDocumentGrant(ctx, request, response, r.client)
}

func (r *workbookGrantResource) Update(context.Context, resource.UpdateRequest, *resource.UpdateResponse) {
}

func (r *workbookGrantResource) Delete(ctx context.Context, request resource.DeleteRequest, response *resource.DeleteResponse) {
	deleteDocumentGrant(ctx, request, response, r.client, "workbooks")
}

func (r *workbookGrantResource) ImportState(ctx context.Context, request resource.ImportStateRequest, response *resource.ImportStateResponse) {
	importGrantCompositeID(ctx, request, response)
}
