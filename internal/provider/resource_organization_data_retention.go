package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &organizationDataRetentionResource{}
	_ resource.ResourceWithConfigure   = &organizationDataRetentionResource{}
	_ resource.ResourceWithImportState = &organizationDataRetentionResource{}
)

type organizationDataRetentionResource struct{ client client.AdminClient }

type organizationDataRetentionResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Type types.String `tfsdk:"type"`
}

func NewOrganizationDataRetentionResource() resource.Resource {
	return &organizationDataRetentionResource{}
}

func (r *organizationDataRetentionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_data_retention"
}

func (r *organizationDataRetentionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization data retention controls. Create and update both call OpenAI's update data-retention endpoint; destroy removes the Terraform object from state because OpenAI does not expose a delete/reset endpoint for organization data retention.",
		Attributes: map[string]resourceschema.Attribute{
			"id":   resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Synthetic singleton ID. Always organization."},
			"type": resourceschema.StringAttribute{Required: true, MarkdownDescription: "Organization data retention type.", Validators: []validator.String{newStringEnumValidator(organizationDataRetentionTypes...)}},
		},
	}
}

func (r *organizationDataRetentionResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationDataRetentionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationDataRetentionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	retention, err := r.client.UpdateOrganizationDataRetention(ctx, client.DataRetentionUpdateRequest{Type: plan.Type.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI organization data retention", err)
		return
	}
	state := organizationDataRetentionResourceModelFromAPI(retention)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationDataRetentionResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	retention, err := r.client.GetOrganizationDataRetention(ctx)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization data retention", err)
		return
	}
	state := organizationDataRetentionResourceModelFromAPI(retention)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationDataRetentionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationDataRetentionResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	retention, err := r.client.UpdateOrganizationDataRetention(ctx, client.DataRetentionUpdateRequest{Type: plan.Type.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to update OpenAI organization data retention", err)
		return
	}
	state := organizationDataRetentionResourceModelFromAPI(retention)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationDataRetentionResource) Delete(ctx context.Context, _ resource.DeleteRequest, resp *resource.DeleteResponse) {
	resp.Diagnostics.AddWarning("OpenAI organization data retention was not reset", "OpenAI exposes retrieve and update for organization data retention but no delete/reset endpoint. Terraform is removing only its state object.")
	resp.State.RemoveResource(ctx)
}

func (r *organizationDataRetentionResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func organizationDataRetentionResourceModelFromAPI(retention *client.DataRetention) organizationDataRetentionResourceModel {
	return organizationDataRetentionResourceModel{ID: types.StringValue(organizationSingletonID), Type: stringOrNull(retention.Type)}
}
