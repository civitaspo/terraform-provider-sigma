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

type workbookScheduleResource struct{ configuredResource }

var (
	_ resource.Resource                = (*workbookScheduleResource)(nil)
	_ resource.ResourceWithConfigure   = (*workbookScheduleResource)(nil)
	_ resource.ResourceWithImportState = (*workbookScheduleResource)(nil)
)

type workbookScheduleModel struct {
	ID         types.String `tfsdk:"id"`
	WorkbookID types.String `tfsdk:"workbook_id"`
	ConfigJSON types.String `tfsdk:"config_json"`
}

func NewWorkbookScheduleResource() resource.Resource { return &workbookScheduleResource{} }

func (r *workbookScheduleResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_workbook_schedule"
}
func (r *workbookScheduleResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	r.configure(req, resp)
}
func (r *workbookScheduleResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Sigma workbook export schedule. Schedule payloads are polymorphic (`target`, `schedule`, `configV2`, and optional fields), so configure them with `config_json`. Sigma list responses omit `target`, so Terraform preserves `config_json` from configuration after create.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{Computed: true, MarkdownDescription: "Scheduled notification ID."},
			"workbook_id": schema.StringAttribute{
				Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
				MarkdownDescription: "Workbook ID.",
			},
			"config_json": schema.StringAttribute{Required: true, MarkdownDescription: "JSON body accepted by the workbook schedule create/update API."},
		},
	}
}
func (r *workbookScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan workbookScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := rawJSON(plan.ConfigJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid workbook schedule config_json", err.Error())
		return
	}
	value, err := r.client.CreateWorkbookSchedule(ctx, plan.WorkbookID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Sigma workbook schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *workbookScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state workbookScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.GetWorkbookSchedule(ctx, state.WorkbookID.ValueString(), state.ID.ValueString())
	if sigma.IsNotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("Unable to read Sigma workbook schedule", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *workbookScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan workbookScheduleModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body, err := rawJSON(plan.ConfigJSON)
	if err != nil {
		resp.Diagnostics.AddError("Invalid workbook schedule config_json", err.Error())
		return
	}
	value, err := r.client.UpdateWorkbookSchedule(ctx, plan.WorkbookID.ValueString(), plan.ID.ValueString(), body)
	if err != nil {
		resp.Diagnostics.AddError("Unable to update Sigma workbook schedule", err.Error())
		return
	}
	plan.ID = types.StringValue(value.ScheduledNotificationID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
func (r *workbookScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state workbookScheduleModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteWorkbookSchedule(ctx, state.WorkbookID.ValueString(), state.ID.ValueString()); err != nil && !sigma.IsNotFound(err) {
		resp.Diagnostics.AddError("Unable to delete Sigma workbook schedule", err.Error())
	}
}
func (r *workbookScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError("Invalid import ID", "Use `workbookId/scheduleId`.")
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workbook_id"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), parts[1])...)
}
