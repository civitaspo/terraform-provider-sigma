package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*userAttributesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*userAttributesDataSource)(nil)
)

type userAttributesDataSource struct{ configuredDataSource }

type userAttributeDataModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Description  types.String `tfsdk:"description"`
	DefaultValue types.String `tfsdk:"default_value"`
}

type userAttributesDataModel struct {
	ID             types.String             `tfsdk:"id"`
	UserAttributes []userAttributeDataModel `tfsdk:"user_attributes"`
}

func NewUserAttributesDataSource() datasource.DataSource { return &userAttributesDataSource{} }

func (d *userAttributesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_attributes"
}

func (d *userAttributesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *userAttributesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma user attributes.", Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"user_attributes": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "User attributes.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "User attribute ID."}, "name": schema.StringAttribute{Computed: true, MarkdownDescription: "Name."},
			"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Description."}, "default_value": schema.StringAttribute{Computed: true, MarkdownDescription: "Default string value."},
		}}},
	}}
}

func (d *userAttributesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListUserAttributes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma user attributes", err.Error())
		return
	}
	state := userAttributesDataModel{ID: types.StringValue("user-attributes")}
	for i := range values {
		state.UserAttributes = append(state.UserAttributes, userAttributeData(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func userAttributeData(v *sigma.UserAttribute) userAttributeDataModel {
	d, x := types.StringNull(), types.StringNull()
	if v.Description != nil {
		d = types.StringValue(*v.Description)
	}
	if v.DefaultValue != nil {
		x = types.StringValue(v.DefaultValue.Val)
	}
	return userAttributeDataModel{ID: types.StringValue(v.UserAttributeID), Name: types.StringValue(v.Name), Description: d, DefaultValue: x}
}
