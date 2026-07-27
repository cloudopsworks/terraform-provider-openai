package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &organizationGroupResource{}
	_ resource.ResourceWithConfigure   = &organizationGroupResource{}
	_ resource.ResourceWithImportState = &organizationGroupResource{}
)

type organizationGroupResource struct{ client client.AdminClient }

type organizationGroupResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	GroupType     types.String `tfsdk:"group_type"`
	IsScimManaged types.Bool   `tfsdk:"is_scim_managed"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
}

func NewOrganizationGroupResource() resource.Resource { return &organizationGroupResource{} }

func (r *organizationGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_group"
}

func (r *organizationGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization group managed through the Administration API. Destroy deletes the group from the organization.",
		Attributes: map[string]resourceschema.Attribute{
			"id":              resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI organization group ID."},
			"name":            resourceschema.StringAttribute{Required: true, MarkdownDescription: "Group display name."},
			"group_type":      resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Group type returned by OpenAI, such as group or tenant_group."},
			"is_scim_managed": resourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the group is managed through SCIM."},
			"created_at":      resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for group creation."},
		},
	}
}

func (r *organizationGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.CreateOrganizationGroup(ctx, client.OrganizationGroupCreateRequest{Name: plan.Name.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI organization group", err)
		return
	}
	state := organizationGroupResourceModelFromAPI(group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.GetOrganizationGroup(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization group", err)
		return
	}
	newState := organizationGroupResourceModelFromAPI(group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationGroupResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	group, err := r.client.UpdateOrganizationGroup(ctx, plan.ID.ValueString(), client.OrganizationGroupUpdateRequest{Name: plan.Name.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI organization group", err)
		return
	}
	state := organizationGroupResourceModelFromAPI(group)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationGroupResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteOrganizationGroup(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI organization group", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *organizationGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func organizationGroupResourceModelFromAPI(group *client.OrganizationGroup) organizationGroupResourceModel {
	return organizationGroupResourceModel{ID: types.StringValue(group.ID), Name: types.StringValue(group.Name), GroupType: stringOrNull(group.GroupType), IsScimManaged: types.BoolValue(group.IsScimManaged), CreatedAt: int64OrNull(group.CreatedAt)}
}
