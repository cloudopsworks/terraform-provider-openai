package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &projectModelPermissionsResource{}
	_ resource.ResourceWithConfigure   = &projectModelPermissionsResource{}
	_ resource.ResourceWithImportState = &projectModelPermissionsResource{}
)
var projectModelPermissionModes = []string{"allow_list", "deny_list"}

type projectModelPermissionsResource struct{ client client.AdminClient }
type projectModelPermissionsResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Mode      types.String `tfsdk:"mode"`
	ModelIDs  types.Set    `tfsdk:"model_ids"`
}

func NewProjectModelPermissionsResource() resource.Resource {
	return &projectModelPermissionsResource{}
}
func (r *projectModelPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_model_permissions"
}
func (r *projectModelPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{MarkdownDescription: "OpenAI project model permissions. Destroy deletes the project model permissions.", Attributes: map[string]resourceschema.Attribute{
		"id":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID, equal to project_id."},
		"project_id": resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"mode":       resourceschema.StringAttribute{Required: true, MarkdownDescription: "Whether the listed models are allowed or denied.", Validators: []validator.String{newStringEnumValidator(projectModelPermissionModes...)}},
		"model_ids":  resourceschema.SetAttribute{Required: true, ElementType: types.StringType, MarkdownDescription: "Model IDs included in the allowlist or denylist policy."},
	}}
}
func (r *projectModelPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, err := configureClient(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("OpenAI provider not configured", err.Error())
		return
	}
	r.client = data.client
}
func (r *projectModelPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectModelPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, diags := r.update(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, diags := projectModelPermissionsResourceModelFromAPI(ctx, permissions, plan.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectModelPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectModelPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, err := r.client.GetProjectModelPermissions(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project model permissions", err)
		return
	}
	state, diags := projectModelPermissionsResourceModelFromAPI(ctx, permissions, state.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectModelPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectModelPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, diags := r.update(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state, diags := projectModelPermissionsResourceModelFromAPI(ctx, permissions, plan.ProjectID)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectModelPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectModelPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteProjectModelPermissions(ctx, state.ProjectID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI project model permissions", err)
		return
	}
	resp.State.RemoveResource(ctx)
}
func (r *projectModelPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}
func (r *projectModelPermissionsResource) update(ctx context.Context, plan projectModelPermissionsResourceModel) (*client.ModelPermissions, diag.Diagnostics) {
	modelIDs, diags := setToSortedStringSlice(ctx, plan.ModelIDs)
	if diags.HasError() {
		return nil, diags
	}
	permissions, err := r.client.UpdateProjectModelPermissions(ctx, plan.ProjectID.ValueString(), client.ModelPermissionsUpdateRequest{Mode: plan.Mode.ValueString(), ModelIDs: modelIDs})
	if err != nil {
		addClientError(&diags, "Unable to update OpenAI project model permissions", err)
	}
	return permissions, diags
}
func projectModelPermissionsResourceModelFromAPI(ctx context.Context, permissions *client.ModelPermissions, projectID types.String) (projectModelPermissionsResourceModel, diag.Diagnostics) {
	modelIDs, diags := setStringValue(ctx, permissions.ModelIDs)
	return projectModelPermissionsResourceModel{ID: projectID, ProjectID: projectID, Mode: stringOrNull(permissions.Mode), ModelIDs: modelIDs}, diags
}
