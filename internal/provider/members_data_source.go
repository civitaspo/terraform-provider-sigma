package provider

import (
	"context"

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
	ID      types.String      `tfsdk:"id"`
	Members []memberDataModel `tfsdk:"members"`
}

func NewMembersDataSource() datasource.DataSource { return &membersDataSource{} }

func (d *membersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_members"
}

func (d *membersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *membersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists all Sigma members.", Attributes: map[string]schema.Attribute{
		"id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"members": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Members.", NestedObject: schema.NestedAttributeObject{Attributes: memberDataAttributes(false)}},
	}}
}

func (d *membersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListMembers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma members", err.Error())
		return
	}
	state := membersDataModel{ID: types.StringValue("members")}
	for i := range values {
		state.Members = append(state.Members, memberData(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
