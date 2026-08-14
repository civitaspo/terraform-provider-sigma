package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*membersDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*membersDataSource)(nil)
)

type membersDataSource struct{ configuredDataSource }

type membersDataModel struct {
	ID              types.String      `tfsdk:"id"`
	Search          types.String      `tfsdk:"search"`
	Email           types.String      `tfsdk:"email"`
	IncludeArchived types.Bool        `tfsdk:"include_archived"`
	IncludeInactive types.Bool        `tfsdk:"include_inactive"`
	Members         []memberDataModel `tfsdk:"members"`
}

func NewMembersDataSource() datasource.DataSource { return &membersDataSource{} }

func (d *membersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_members"
}

func (d *membersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *membersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma members with documented list filters." + listCollectionNotice, Attributes: map[string]schema.Attribute{
		"id":               schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"search":           schema.StringAttribute{Optional: true, MarkdownDescription: "Search filter (`search`)."},
		"email":            schema.StringAttribute{Optional: true, MarkdownDescription: "Exact email filter (`email`)."},
		"include_archived": schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether archived members are included (`includeArchived`). Explicit `false` is sent; null omits the parameter."},
		"include_inactive": schema.BoolAttribute{Optional: true, MarkdownDescription: "Whether inactive SCIM members are included (`includeInactive`). Explicit `false` is sent; null omits the parameter."},
		"members":          schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Members in API order.", NestedObject: schema.NestedAttributeObject{Attributes: memberDataAttributes(false)}},
	}}
}

func (d *membersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state membersDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if abortUnknownInputs(&resp.Diagnostics, state.Search, state.Email, state.IncludeArchived, state.IncludeInactive) {
		return
	}
	values, err := d.client.ListMembers(ctx, sigma.ListMembersOptions{
		Search:          optionalStringPtr(state.Search),
		Email:           optionalStringPtr(state.Email),
		IncludeArchived: optionalBoolPtr(state.IncludeArchived),
		IncludeInactive: optionalBoolPtr(state.IncludeInactive),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma members", err.Error())
		return
	}
	state.ID = types.StringValue("members")
	state.Members = make([]memberDataModel, 0, len(values))
	for i := range values {
		state.Members = append(state.Members, memberData(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
