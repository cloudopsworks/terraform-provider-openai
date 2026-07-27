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
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

type projectResource struct {
	client client.AdminClient
}

type projectResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	ExternalKeyID types.String `tfsdk:"external_key_id"`
	Geography     types.String `tfsdk:"geography"`
	Status        types.String `tfsdk:"status"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	ArchivedAt    types.Int64  `tfsdk:"archived_at"`
}

func NewProjectResource() resource.Resource { return &projectResource{} }

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization project managed through the Administration API. Destroy archives the project because OpenAI projects are archived rather than hard-deleted.",
		Attributes: map[string]resourceschema.Attribute{
			"id":              resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI project ID."},
			"name":            resourceschema.StringAttribute{Required: true, MarkdownDescription: "Project display name."},
			"external_key_id": resourceschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Optional external key ID associated with the project."},
			"geography":       resourceschema.StringAttribute{Optional: true, MarkdownDescription: "Optional data residency geography for project creation/update when enabled for the organization.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"status":          resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project status, such as active or archived."},
			"created_at":      resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for project creation."},
			"archived_at":     resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for project archival when archived."},
		},
	}
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, err := r.client.CreateProject(ctx, client.ProjectCreateRequest{Name: plan.Name.ValueString(), ExternalKeyID: stringValue(plan.ExternalKeyID), Geography: stringValue(plan.Geography)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI project", err)
		return
	}
	state := projectModelFromAPI(project, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, err := r.client.GetProject(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI project", err)
		return
	}
	newState := projectModelFromAPI(project, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	project, err := r.client.UpdateProject(ctx, plan.ID.ValueString(), client.ProjectUpdateRequest{Name: plan.Name.ValueString(), ExternalKeyID: stringValue(plan.ExternalKeyID), Geography: stringValue(plan.Geography)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI project", err)
		return
	}
	state := projectModelFromAPI(project, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	_, err := r.client.ArchiveProject(ctx, state.ID.ValueString())
	if err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to archive OpenAI project", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func projectModelFromAPI(project *client.Project, prior projectResourceModel) projectResourceModel {
	return projectResourceModel{
		ID:            types.StringValue(project.ID),
		Name:          types.StringValue(project.Name),
		ExternalKeyID: stringOrNull(project.ExternalKeyID),
		Geography:     prior.Geography,
		Status:        stringOrNull(project.Status),
		CreatedAt:     int64OrNull(project.CreatedAt),
		ArchivedAt:    int64OrNull(project.ArchivedAt),
	}
}
