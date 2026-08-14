package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
)

var (
	_ datasource.DataSource              = (*userAttributeDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*userAttributeDataSource)(nil)
)

type userAttributeDataSource struct{ configuredDataSource }

type userAttributeLookupModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	DefaultValue types.String `tfsdk:"default_value"`
	CreatedBy    types.String `tfsdk:"created_by"`
	UpdatedBy    types.String `tfsdk:"updated_by"`
	CreatedAt    types.String `tfsdk:"created_at"`
	UpdatedAt    types.String `tfsdk:"updated_at"`
}

func NewUserAttributeDataSource() datasource.DataSource { return &userAttributeDataSource{} }

func (d *userAttributeDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attribute"
}

func (d *userAttributeDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *userAttributeDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Retrieves a Sigma user attribute by ID.",
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Required: true, MarkdownDescription: "User attribute ID."},
			"name":          schema.StringAttribute{Computed: true, MarkdownDescription: "User attribute name."},
			"description":   schema.StringAttribute{Computed: true, MarkdownDescription: "User attribute description."},
			"default_value": schema.StringAttribute{Computed: true, MarkdownDescription: "Default string value."},
			"created_by":    schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID that created the user attribute."},
			"updated_by":    schema.StringAttribute{Computed: true, MarkdownDescription: "Member ID that last updated the user attribute."},
			"created_at":    schema.StringAttribute{Computed: true, MarkdownDescription: "Creation timestamp."},
			"updated_at":    schema.StringAttribute{Computed: true, MarkdownDescription: "Last update timestamp."},
		},
	}
}

func (d *userAttributeDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state userAttributeLookupModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetUserAttribute(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma user attribute", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, userAttributeLookup(value))...)
}

func userAttributeLookup(value *sigma.UserAttribute) userAttributeLookupModel {
	nested := userAttributeData(value)
	return userAttributeLookupModel{
		ID: nested.ID, Name: nested.Name, Description: nested.Description, DefaultValue: nested.DefaultValue,
		CreatedBy: types.StringValue(value.CreatedBy), UpdatedBy: types.StringValue(value.UpdatedBy),
		CreatedAt: types.StringValue(value.CreatedAt), UpdatedAt: types.StringValue(value.UpdatedAt),
	}
}
