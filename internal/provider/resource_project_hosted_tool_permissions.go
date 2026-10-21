package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &projectHostedToolPermissionsResource{}
	_ resource.ResourceWithConfigure   = &projectHostedToolPermissionsResource{}
	_ resource.ResourceWithImportState = &projectHostedToolPermissionsResource{}
)

type projectHostedToolPermissionsResource struct{ client client.AdminClient }
type projectHostedToolPermissionsResourceModel struct {
	ID              types.String `tfsdk:"id"`
	ProjectID       types.String `tfsdk:"project_id"`
	CodeInterpreter types.Bool   `tfsdk:"code_interpreter"`
	FileSearch      types.Bool   `tfsdk:"file_search"`
	ImageGeneration types.Bool   `tfsdk:"image_generation"`
	Mcp             types.Bool   `tfsdk:"mcp"`
	WebSearch       types.Bool   `tfsdk:"web_search"`
}

func NewProjectHostedToolPermissionsResource() resource.Resource {
	return &projectHostedToolPermissionsResource{}
}
func (r *projectHostedToolPermissionsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_hosted_tool_permissions"
}
func (r *projectHostedToolPermissionsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{MarkdownDescription: "OpenAI project hosted-tool permissions. Destroy removes only the Terraform state because OpenAI does not expose a delete/reset endpoint.", Attributes: map[string]resourceschema.Attribute{
		"id":               resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID, equal to project_id."},
		"project_id":       resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"code_interpreter": resourceschema.BoolAttribute{Required: true, MarkdownDescription: "Whether Code Interpreter is enabled."},
		"file_search":      resourceschema.BoolAttribute{Required: true, MarkdownDescription: "Whether File Search is enabled."},
		"image_generation": resourceschema.BoolAttribute{Required: true, MarkdownDescription: "Whether Image Generation is enabled."},
		"mcp":              resourceschema.BoolAttribute{Required: true, MarkdownDescription: "Whether MCP is enabled."},
		"web_search":       resourceschema.BoolAttribute{Required: true, MarkdownDescription: "Whether Web Search is enabled."},
	}}
}
func (r *projectHostedToolPermissionsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (r *projectHostedToolPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectHostedToolPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, err := r.update(ctx, plan)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project hosted-tool permissions", err)
		return
	}
	state := projectHostedToolPermissionsResourceModelFromAPI(permissions, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectHostedToolPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectHostedToolPermissionsResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, err := r.client.GetProjectHostedToolPermissions(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project hosted-tool permissions", err)
		return
	}
	state = projectHostedToolPermissionsResourceModelFromAPI(permissions, state.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectHostedToolPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectHostedToolPermissionsResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	permissions, err := r.update(ctx, plan)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project hosted-tool permissions", err)
		return
	}
	state := projectHostedToolPermissionsResourceModelFromAPI(permissions, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
func (r *projectHostedToolPermissionsResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("OpenAI project hosted-tool permissions were not reset", "OpenAI exposes retrieve and update for project hosted-tool permissions but no delete/reset endpoint. Terraform is removing only its state object.")
	resp.State.RemoveResource(ctx)
}
func (r *projectHostedToolPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}
func (r *projectHostedToolPermissionsResource) update(ctx context.Context, plan projectHostedToolPermissionsResourceModel) (*client.HostedToolPermissions, error) {
	return r.client.UpdateProjectHostedToolPermissions(ctx, plan.ProjectID.ValueString(), client.HostedToolPermissionsUpdateRequest{CodeInterpreter: plan.CodeInterpreter.ValueBool(), FileSearch: plan.FileSearch.ValueBool(), ImageGeneration: plan.ImageGeneration.ValueBool(), Mcp: plan.Mcp.ValueBool(), WebSearch: plan.WebSearch.ValueBool()})
}
func projectHostedToolPermissionsResourceModelFromAPI(permissions *client.HostedToolPermissions, projectID types.String) projectHostedToolPermissionsResourceModel {
	return projectHostedToolPermissionsResourceModel{ID: projectID, ProjectID: projectID, CodeInterpreter: types.BoolValue(permissions.CodeInterpreter), FileSearch: types.BoolValue(permissions.FileSearch), ImageGeneration: types.BoolValue(permissions.ImageGeneration), Mcp: types.BoolValue(permissions.Mcp), WebSearch: types.BoolValue(permissions.WebSearch)}
}
