package provider

import (
	"context"

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
	_ resource.Resource                = &projectDataRetentionResource{}
	_ resource.ResourceWithConfigure   = &projectDataRetentionResource{}
	_ resource.ResourceWithImportState = &projectDataRetentionResource{}
)

var projectDataRetentionTypes = []string{
	"organization_default",
	"none",
	"zero_data_retention",
	"modified_abuse_monitoring",
	"enhanced_zero_data_retention",
	"enhanced_modified_abuse_monitoring",
}

type projectDataRetentionResource struct{ client client.AdminClient }

type projectDataRetentionResourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Type      types.String `tfsdk:"type"`
}

func NewProjectDataRetentionResource() resource.Resource { return &projectDataRetentionResource{} }

func (r *projectDataRetentionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_data_retention"
}

func (r *projectDataRetentionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{MarkdownDescription: "OpenAI project data retention controls. Destroy removes only the Terraform state because OpenAI does not expose a delete/reset endpoint.", Attributes: map[string]resourceschema.Attribute{
		"id":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID, equal to project_id."},
		"project_id": resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"type":       resourceschema.StringAttribute{Required: true, MarkdownDescription: "Project data retention type.", Validators: []validator.String{newStringEnumValidator(projectDataRetentionTypes...)}},
	}}
}

func (r *projectDataRetentionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectDataRetentionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectDataRetentionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	retention, err := r.client.UpdateProjectDataRetention(ctx, plan.ProjectID.ValueString(), client.DataRetentionUpdateRequest{Type: plan.Type.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project data retention", err)
		return
	}
	state := projectDataRetentionResourceModelFromAPI(retention, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectDataRetentionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectDataRetentionResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	retention, err := r.client.GetProjectDataRetention(ctx, state.ProjectID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project data retention", err)
		return
	}
	state = projectDataRetentionResourceModelFromAPI(retention, state.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectDataRetentionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectDataRetentionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	retention, err := r.client.UpdateProjectDataRetention(ctx, plan.ProjectID.ValueString(), client.DataRetentionUpdateRequest{Type: plan.Type.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project data retention", err)
		return
	}
	state := projectDataRetentionResourceModelFromAPI(retention, plan.ProjectID)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectDataRetentionResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("OpenAI project data retention was not reset", "OpenAI exposes retrieve and update for project data retention but no delete/reset endpoint. Terraform is removing only its state object.")
	resp.State.RemoveResource(ctx)
}

func (r *projectDataRetentionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), types.StringValue(req.ID))...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), types.StringValue(req.ID))...)
}

func projectDataRetentionResourceModelFromAPI(retention *client.DataRetention, projectID types.String) projectDataRetentionResourceModel {
	return projectDataRetentionResourceModel{ID: projectID, ProjectID: projectID, Type: stringOrNull(retention.Type)}
}
