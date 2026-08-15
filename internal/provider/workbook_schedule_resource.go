package provider

import (
	"context"

	"github.com/civitaspo/terraform-provider-sigma/internal/sigma"
	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type workbookScheduleResource struct{ configuredResource }

var (
	_ resource.Resource                   = (*workbookScheduleResource)(nil)
	_ resource.ResourceWithConfigure      = (*workbookScheduleResource)(nil)
	_ resource.ResourceWithValidateConfig = (*workbookScheduleResource)(nil)
)

type workbookScheduleModel struct {
	ID          types.String         `tfsdk:"id"`
	WorkbookID  types.String         `tfsdk:"workbook_id"`
	ConfigJSON  jsontypes.Normalized `tfsdk:"config_json"`
	IsSuspended types.Bool           `tfsdk:"is_suspended"`
}

func NewWorkbookScheduleResource() resource.Resource { return &workbookScheduleResource{} }

func (r *workbookScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workbook_schedule"
}

func (r *workbookScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}

func (r *workbookScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = scheduleSchema(
		"Manages a Sigma workbook export schedule. Schedule payloads are polymorphic (`target`, `schedule`, `configV2`, and optional fields), so configure them with `config_json`. Sigma list responses omit `target`, so Terraform retains request-only fields from prior state. Import is not supported because Read cannot reconstruct `target`.",
		"workbook_id",
		"Workbook ID.",
	)
}

func (r *workbookScheduleResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config workbookScheduleModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(validateScheduleConfigJSON(config.ConfigJSON)...)
}

func (r *workbookScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workbookScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	workbookID, diags := knownString(plan.WorkbookID, "workbook_id")
	resp.Diagnostics.Append(diags...)
	body, bodyDiags := scheduleConfigForRequest(plan.ConfigJSON)
	resp.Diagnostics.Append(bodyDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.CreateWorkbookSchedule(ctx, workbookID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma workbook schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	if patch, needed, patchDiags := applyScheduleCreateSuspension(plan.ConfigJSON, plan.IsSuspended, value.IsSuspended); patchDiags.HasError() {
		resp.Diagnostics.Append(patchDiags...)
		return
	} else if needed {
		value, err = r.client.UpdateWorkbookSchedule(ctx, workbookID, value.ScheduledNotificationID, patch)
		if err != nil {
			resp.Diagnostics.AddError("Unable to pause or resume Sigma workbook schedule", err.Error())
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
	plan.IsSuspended = scheduleIsSuspended(plan.IsSuspended, value.IsSuspended)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workbookScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workbookScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	workbookID, diags := knownString(state.WorkbookID, "workbook_id")
	resp.Diagnostics.Append(diags...)
	scheduleID, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.GetWorkbookSchedule(ctx, workbookID, scheduleID)
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma workbook schedule", err.Error())
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

func (r *workbookScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workbookScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	var state workbookScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	workbookID, diags := knownString(plan.WorkbookID, "workbook_id")
	resp.Diagnostics.Append(diags...)
	scheduleID, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	body, bodyDiags := scheduleUpdateBody(plan.ConfigJSON, plan.IsSuspended, state.IsSuspended)
	resp.Diagnostics.Append(bodyDiags...)
	if resp.Diagnostics.HasError() {
		return
	}
	value, err := r.client.UpdateWorkbookSchedule(ctx, workbookID, scheduleID, body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma workbook schedule", err.Error())
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
	plan.IsSuspended = scheduleIsSuspended(plan.IsSuspended, value.IsSuspended)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *workbookScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workbookScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	workbookID, diags := knownString(state.WorkbookID, "workbook_id")
	resp.Diagnostics.Append(diags...)
	scheduleID, diags := knownString(state.ID, "id")
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWorkbookSchedule(ctx, workbookID, scheduleID); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma workbook schedule", err.Error())
	}
}
