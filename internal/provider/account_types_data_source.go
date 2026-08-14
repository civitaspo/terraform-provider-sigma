package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*accountTypesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*accountTypesDataSource)(nil)
)

type accountTypesDataSource struct{ configuredDataSource }

type accountTypeDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsCustom    types.Bool   `tfsdk:"is_custom"`
}

type accountTypesDataModel struct {
	ID           types.String           `tfsdk:"id"`
	AccountTypes []accountTypeDataModel `tfsdk:"account_types"`
}

func NewAccountTypesDataSource() datasource.DataSource { return &accountTypesDataSource{} }

func (d *accountTypesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_account_types"
}

func (d *accountTypesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *accountTypesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma account types." + listCollectionNotice, Attributes: map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"account_types": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Account types.", NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Account type ID."}, "name": schema.StringAttribute{Computed: true, MarkdownDescription: "Account type name."},
			"description": schema.StringAttribute{Computed: true, MarkdownDescription: "Description."}, "is_custom": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the type is custom."},
		}}},
	}}
}

func (d *accountTypesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListAccountTypes(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma account types", err.Error())
		return
	}
	state := accountTypesDataModel{ID: types.StringValue("account-types"), AccountTypes: make([]accountTypeDataModel, 0, len(values))}
	for _, v := range values {
		state.AccountTypes = append(state.AccountTypes, accountTypeDataModel{ID: types.StringValue(v.AccountTypeID), Name: types.StringValue(v.AccountTypeName), Description: types.StringValue(v.Description), IsCustom: types.BoolValue(v.IsCustom)})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
