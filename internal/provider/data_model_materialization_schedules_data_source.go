package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*dataModelMaterializationSchedulesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*dataModelMaterializationSchedulesDataSource)(nil)
)

type dataModelMaterializationSchedulesDataSource struct{ configuredDataSource }

type dataModelMaterializationSchedulesModel struct {
	ID          types.String                   `tfsdk:"id"`
	DataModelID types.String                   `tfsdk:"data_model_id"`
	Schedules   []materializationScheduleModel `tfsdk:"schedules"`
}

func NewDataModelMaterializationSchedulesDataSource() datasource.DataSource {
	return &dataModelMaterializationSchedulesDataSource{}
}

func (d *dataModelMaterializationSchedulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_data_model_materialization_schedules"
}

func (d *dataModelMaterializationSchedulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *dataModelMaterializationSchedulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists materialization schedules for a Sigma data model." + listCollectionNotice,
		Attributes: map[string]schema.Attribute{
			"id":            schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"data_model_id": schema.StringAttribute{Required: true, MarkdownDescription: "Data model ID."},
			"schedules":     schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Materialization schedules in API order.", NestedObject: schema.NestedAttributeObject{Attributes: materializationScheduleAttributes()}},
		},
	}
}

func (d *dataModelMaterializationSchedulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state dataModelMaterializationSchedulesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.DataModelID, "data_model_id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	values, err := d.client.ListDataModelMaterializationSchedules(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma data model materialization schedules", err.Error())
		return
	}
	state.ID = types.StringValue(id)
	state.Schedules = materializationScheduleData(values)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
