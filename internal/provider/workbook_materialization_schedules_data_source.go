package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = (*workbookMaterializationSchedulesDataSource)(nil)
	_ datasource.DataSourceWithConfigure = (*workbookMaterializationSchedulesDataSource)(nil)
)

type workbookMaterializationSchedulesDataSource struct{ configuredDataSource }

type materializationCronModel struct {
	CronSpec types.String `tfsdk:"cron_spec"`
	Timezone types.String `tfsdk:"timezone"`
}

type materializationScheduleModel struct {
	SheetID      types.String              `tfsdk:"sheet_id"`
	ElementID    types.String              `tfsdk:"element_id"`
	ElementName  types.String              `tfsdk:"element_name"`
	Schedule     *materializationCronModel `tfsdk:"schedule"`
	ConfiguredAt types.String              `tfsdk:"configured_at"`
	Paused       types.Bool                `tfsdk:"paused"`
}

type workbookMaterializationSchedulesModel struct {
	ID         types.String                   `tfsdk:"id"`
	WorkbookID types.String                   `tfsdk:"workbook_id"`
	Schedules  []materializationScheduleModel `tfsdk:"schedules"`
}

func NewWorkbookMaterializationSchedulesDataSource() datasource.DataSource {
	return &workbookMaterializationSchedulesDataSource{}
}

func (d *workbookMaterializationSchedulesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workbook_materialization_schedules"
}

func (d *workbookMaterializationSchedulesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.configure(req, resp)
}

func (d *workbookMaterializationSchedulesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists materialization schedules for a Sigma workbook using the paginated v2.1 endpoint." + listCollectionNotice,
		Attributes: map[string]schema.Attribute{
			"id":          schema.StringAttribute{Computed: true, MarkdownDescription: "Stable identifier for this data source."},
			"workbook_id": schema.StringAttribute{Required: true, MarkdownDescription: "Workbook ID."},
			"schedules":   schema.ListNestedAttribute{Computed: true, MarkdownDescription: "Materialization schedules in API order.", NestedObject: schema.NestedAttributeObject{Attributes: materializationScheduleAttributes()}},
		},
	}
}

func (d *workbookMaterializationSchedulesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state workbookMaterializationSchedulesModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id, idDiags := knownString(state.WorkbookID, "workbook_id")
	resp.Diagnostics.Append(idDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	values, err := d.client.ListWorkbookMaterializationSchedules(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Unable to list Sigma workbook materialization schedules", err.Error())
		return
	}
	state.ID = types.StringValue(id)
	state.Schedules = materializationScheduleData(values)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func materializationScheduleAttributes() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"sheet_id":      schema.StringAttribute{Computed: true, MarkdownDescription: "Identifier of the materialization schedule for the element (`sheetId`)."},
		"element_id":    schema.StringAttribute{Computed: true, MarkdownDescription: "Identifier of the element being materialized (`elementId`)."},
		"element_name":  schema.StringAttribute{Computed: true, MarkdownDescription: "Element name (`elementName`)."},
		"configured_at": schema.StringAttribute{Computed: true, MarkdownDescription: "When the schedule was configured (`configuredAt`)."},
		"paused":        schema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the schedule is paused (`paused`)."},
		"schedule": schema.SingleNestedAttribute{Computed: true, MarkdownDescription: "Cron schedule (`schedule`). Null when Sigma returns no schedule.", Attributes: map[string]schema.Attribute{
			"cron_spec": schema.StringAttribute{Computed: true, MarkdownDescription: "Cron expression (`cronSpec`)."},
			"timezone":  schema.StringAttribute{Computed: true, MarkdownDescription: "Timezone code (`timezone`)."},
		}},
	}
}

func materializationScheduleData(values []sigma.MaterializationSchedule) []materializationScheduleModel {
	items := make([]materializationScheduleModel, 0, len(values))
	for _, value := range values {
		item := materializationScheduleModel{
			SheetID: types.StringValue(value.SheetID), ElementID: types.StringValue(value.ElementID),
			ElementName: types.StringValue(value.ElementName), ConfiguredAt: types.StringValue(value.ConfiguredAt),
			Paused: types.BoolValue(value.Paused),
		}
		if value.Schedule != nil {
			item.Schedule = &materializationCronModel{
				CronSpec: types.StringValue(value.Schedule.CronSpec),
				Timezone: types.StringValue(value.Schedule.Timezone),
			}
		}
		items = append(items, item)
	}
	return items
}
