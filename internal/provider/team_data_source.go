package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*teamDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*teamDataSource)(nil)
)

type teamDataSource struct{ configuredDataSource }

type teamDataModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	Visibility  types.String `tfsdk:"visibility"`
	IsArchived  types.Bool   `tfsdk:"is_archived"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
}

func NewTeamDataSource() datasource.DataSource { return &teamDataSource{} }

func (d *teamDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_team"
}

func (d *teamDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *teamDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{MarkdownDescription: "Retrieves a Sigma team by ID.", Attributes: teamDataAttributes(true)}
}

func (d *teamDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state teamDataModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.ID, "id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := d.client.GetTeam(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma team", err.Error())
		return
	}
	state = teamData(value)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func teamDataAttributes(requireID bool) map[string]schema.Attribute {
	id := schema.StringAttribute{Computed: true, MarkdownDescription: "Team ID."}
	if requireID {
		id = schema.StringAttribute{Required: true, MarkdownDescription: "Team ID."}
	}
	return map[string]schema.Attribute{
		"id": id, "name": schema.StringAttribute{Computed: true, MarkdownDescription: "Team name."}, "description": schema.StringAttribute{Computed: true, MarkdownDescription: "Description."},
		"visibility": schema.StringAttribute{Computed: true, MarkdownDescription: "Visibility."}, "is_archived": schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether archived."},
		"workspace_id": schema.StringAttribute{Computed: true, MarkdownDescription: "ID of the team workspace when the API returns it."},
	}
}

func teamData(v *sigma.Team) teamDataModel {
	d := types.StringNull()
	if v.Description != nil {
		d = types.StringValue(*v.Description)
	}
	workspaceID := types.StringNull()
	if v.WorkspaceID != nil && *v.WorkspaceID != "" {
		workspaceID = types.StringValue(*v.WorkspaceID)
	}
	return teamDataModel{ID: types.StringValue(v.TeamID), Name: types.StringValue(v.Name), Description: d, Visibility: types.StringValue(v.Visibility), IsArchived: types.BoolValue(v.IsArchived), WorkspaceID: workspaceID}
}
