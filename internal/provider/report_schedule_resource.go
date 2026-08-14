package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type reportScheduleResource struct{ configuredResource }

var (
	_ resource.Resource                   = (*reportScheduleResource)(nil)
	_ resource.ResourceWithConfigure      = (*reportScheduleResource)(nil)
	_ resource.ResourceWithValidateConfig = (*reportScheduleResource)(nil)
)

type reportScheduleModel struct {
	ID          types.String         `tfsdk:"id"`
	ReportID    types.String         `tfsdk:"report_id"`
	ConfigJSON  jsontypes.Normalized `tfsdk:"config_json"`
	IsSuspended types.Bool           `tfsdk:"is_suspended"`
}

func NewReportScheduleResource() resource.Resource { return &reportScheduleResource{} }

func (r *reportScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_report_schedule"
}

func (r *reportScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *reportScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = scheduleSchema(
		"Manages a Sigma report export schedule. Schedule payloads are polymorphic, so configure them with `config_json`. Sigma list responses omit `target`, so Terraform retains request-only fields from prior state. Import is not supported because Read cannot reconstruct `target`.",
		"report_id",
		"Report ID.",
	)
}

func (r *reportScheduleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config reportScheduleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateScheduleConfigJSON(config.ConfigJSON)...)
}

func (r *reportScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan reportScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	reportID, diags := knownString(plan.ReportID, "report_id")
	resp.Diagnostics.Append(diags...)
	body, bodyDiags := scheduleConfigForRequest(plan.ConfigJSON)
	resp.Diagnostics.Append(bodyDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateReportSchedule(ctx, reportID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma report schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	if patch, needed, patchErr := applyScheduleCreateSuspension(plan.IsSuspended, value.IsSuspended); patchErr != nil {
		resp.Diagnostics.AddError("Unable to pause or resume Sigma report schedule", patchErr.Error())
		return
	} else if needed {
		value, err = r.client.UpdateReportSchedule(ctx, reportID, value.ScheduledNotificationID, patch)
		if err != nil {
			resp.Diagnostics.AddError("Unable to pause or resume Sigma report schedule", err.Error())
			return
		}
	}
	config, mergeDiags := mergeScheduleConfig(plan.ConfigJSON, scheduleRefresh{
		Schedule: value.Schedule, ConfigV2: value.ConfigV2, IsSuspended: value.IsSuspended,
	})
	resp.Diagnostics.Append(mergeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ConfigJSON = config
	plan.IsSuspended = types.BoolValue(value.IsSuspended)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reportScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state reportScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	reportID, diags := knownString(state.ReportID, "report_id")
	resp.Diagnostics.Append(diags...)
	scheduleID, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetReportSchedule(ctx, reportID, scheduleID)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma report schedule", err.Error())
		return
	}
	config, mergeDiags := mergeScheduleConfig(state.ConfigJSON, scheduleRefresh{
		Schedule: value.Schedule, ConfigV2: value.ConfigV2, IsSuspended: value.IsSuspended,
	})
	resp.Diagnostics.Append(mergeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.ConfigJSON = config
	state.IsSuspended = types.BoolValue(value.IsSuspended)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *reportScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan reportScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state reportScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	reportID, diags := knownString(plan.ReportID, "report_id")
	resp.Diagnostics.Append(diags...)
	scheduleID, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	body, bodyDiags := scheduleUpdateBody(plan.ConfigJSON, plan.IsSuspended, state.IsSuspended)
	resp.Diagnostics.Append(bodyDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.UpdateReportSchedule(ctx, reportID, scheduleID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma report schedule", err.Error())
		return
	}
	config, mergeDiags := mergeScheduleConfig(plan.ConfigJSON, scheduleRefresh{
		Schedule: value.Schedule, ConfigV2: value.ConfigV2, IsSuspended: value.IsSuspended,
	})
	resp.Diagnostics.Append(mergeDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	plan.ConfigJSON = config
	plan.IsSuspended = types.BoolValue(value.IsSuspended)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *reportScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state reportScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	reportID, diags := knownString(state.ReportID, "report_id")
	resp.Diagnostics.Append(diags...)
	scheduleID, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteReportSchedule(ctx, reportID, scheduleID); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma report schedule", err.Error())
	}
}
