package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*teamsDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*teamsDataSource)(nil)
)

type teamsDataSource struct{ configuredDataSource }

type teamsDataModel struct {
	ID    types.String    `tfsdk:"id"`
	Teams []teamDataModel `tfsdk:"teams"`
}

func NewTeamsDataSource() datasource.DataSource { return &teamsDataSource{} }

func (d *teamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *teamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *teamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists all Sigma teams using the paginated v2.1 endpoint.", Attributes: map[string]schema.Attribute{
		"id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"teams": schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Teams.", NestedObject: schema.NestedAttributeObject{Attributes: teamDataAttributes(false)}},
	}}
}

func (d *teamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	values, err := d.client.ListTeams(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma teams", err.Error())
		return
	}
	state := teamsDataModel{ID: types.StringValue("teams")}
	for i := range values {
		state.Teams = append(state.Teams, teamData(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
