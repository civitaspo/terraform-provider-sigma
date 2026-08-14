package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
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
	ID          types.String    `tfsdk:"id"`
	Name        types.String    `tfsdk:"name"`
	Description types.String    `tfsdk:"description"`
	Visibility  types.String    `tfsdk:"visibility"`
	Teams       []teamDataModel `tfsdk:"teams"`
}

func NewTeamsDataSource() datasource.DataSource { return &teamsDataSource{} }

func (d *teamsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_teams"
}

func (d *teamsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *teamsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Lists Sigma teams using the paginated v2.1 endpoint." + listCollectionNotice, Attributes: map[string]schema.Attribute{
		"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
		"name":        schema.StringAttribute{Optional: true, MarkdownDescription: "Name filter (`name`)."},
		"description": schema.StringAttribute{Optional: true, MarkdownDescription: "Description filter (`description`)."},
		"visibility":  schema.StringAttribute{Optional: true, MarkdownDescription: "Visibility filter (`visibility`): `public` or `private`."},
		"teams":       schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Teams in API order.", NestedObject: schema.NestedAttributeObject{Attributes: teamDataAttributes(false)}},
	}}
}

func (d *teamsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state teamsDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if abortUnknownInputs(&resp.Diagnostics, state.Name, state.Description, state.Visibility) {
		return
	}
	values, err := d.client.ListTeams(ctx, sigma.ListTeamsOptions{
		Name:        optionalStringPtr(state.Name),
		Description: optionalStringPtr(state.Description),
		Visibility:  optionalStringPtr(state.Visibility),
	})
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma teams", err.Error())
		return
	}
	state.ID = types.StringValue("teams")
	state.Teams = make([]teamDataModel, 0, len(values))
	for i := range values {
		state.Teams = append(state.Teams, teamData(&values[i]))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
