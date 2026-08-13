package provider

import (
	"context"
	"strings"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type reportScheduleResource struct{ configuredResource }

var (
	_ resource.Resource                = (*reportScheduleResource)(nil)
	_ resource.ResourceWithConfigure   = (*reportScheduleResource)(nil)
	_ resource.ResourceWithImportState = (*reportScheduleResource)(nil)
)

type reportScheduleModel struct {
	ID         types.String `tfsdk:"id"`
	ReportID   types.String `tfsdk:"report_id"`
	ConfigJSON types.String `tfsdk:"config_json"`
}

func NewReportScheduleResource() resource.Resource { return &reportScheduleResource{} }

func (r *reportScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_schedule"
}
func (r *reportScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *reportScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma report export schedule. Schedule payloads are polymorphic, so configure them with `config_json`. Sigma list responses omit `target`, so Terraform preserves `config_json` from configuration after create.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Scheduled notification ID."},
			"report_id": schema.StringAttribute{
				Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Report ID.",
			},
			"config_json": schema.StringAttribute{Required: true, MarkdownDescription: "JSON body accepted by the report schedule create/update API."},
		},
	}
}
func (r *reportScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan reportScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := rawJSON(plan.ConfigJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid report schedule config_json", err.Error())
		return
	}
	value, err := r.client.CreateReportSchedule(ctx, plan.ReportID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma report schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *reportScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state reportScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.GetReportSchedule(ctx, state.ReportID.ValueString(), state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma report schedule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *reportScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan reportScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := rawJSON(plan.ConfigJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid report schedule config_json", err.Error())
		return
	}
	value, err := r.client.UpdateReportSchedule(ctx, plan.ReportID.ValueString(), plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma report schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *reportScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state reportScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteReportSchedule(ctx, state.ReportID.ValueString(), state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma report schedule", err.Error())
	}
}
func (r *reportScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Use `reportId/scheduleId`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("report_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
