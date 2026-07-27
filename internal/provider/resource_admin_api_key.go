package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &adminAPIKeyResource{}
	_ resource.ResourceWithConfigure   = &adminAPIKeyResource{}
	_ resource.ResourceWithImportState = &adminAPIKeyResource{}
)

type adminAPIKeyResource struct {
	client client.AdminClient
}

type adminAPIKeyResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	ExpiresInSeconds types.Int64  `tfsdk:"expires_in_seconds"`
	Value            types.String `tfsdk:"value"`
	RedactedValue    types.String `tfsdk:"redacted_value"`
	OwnerType        types.String `tfsdk:"owner_type"`
	OwnerID          types.String `tfsdk:"owner_id"`
	OwnerName        types.String `tfsdk:"owner_name"`
	OwnerRole        types.String `tfsdk:"owner_role"`
	CreatedAt        types.Int64  `tfsdk:"created_at"`
	ExpiresAt        types.Int64  `tfsdk:"expires_at"`
	LastUsedAt       types.Int64  `tfsdk:"last_used_at"`
}

func NewAdminAPIKeyResource() resource.Resource { return &adminAPIKeyResource{} }

func (r *adminAPIKeyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_api_key"
}

func (r *adminAPIKeyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization admin API key. The unredacted value is returned only on create and stored as a Sensitive computed state attribute; use Terraform lifecycle prevent_destroy for critical admin keys.",
		Attributes: map[string]resourceschema.Attribute{
			"id":                 resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI admin API key ID."},
			"name":               resourceschema.StringAttribute{Required: true, MarkdownDescription: "Admin API key name. OpenAI does not expose update for admin keys, so changes replace the resource.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"expires_in_seconds": resourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Optional number of seconds until the admin API key expires. Omit for a non-expiring key. Changes replace the key.", PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplace()}},
			"value":              resourceschema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Unredacted admin API key value returned only during create. Protect Terraform state accordingly."},
			"redacted_value":     resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Redacted admin API key value returned by read operations."},
			"owner_type":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner type returned by OpenAI."},
			"owner_id":           resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner ID returned by OpenAI."},
			"owner_name":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner display name returned by OpenAI."},
			"owner_role":         resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner role returned by OpenAI."},
			"created_at":         resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for key creation."},
			"expires_at":         resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for key expiration when configured."},
			"last_used_at":       resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the key was last used, if known."},
		},
	}
}

func (r *adminAPIKeyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *adminAPIKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan adminAPIKeyResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey, err := r.client.CreateAdminAPIKey(ctx, client.AdminAPIKeyCreateRequest{Name: plan.Name.ValueString(), ExpiresInSeconds: int64Value(plan.ExpiresInSeconds)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI admin API key", err)
		return
	}
	state := adminAPIKeyResourceModelFromAPI(apiKey, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *adminAPIKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state adminAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey, err := r.client.GetAdminAPIKey(ctx, state.ID.ValueString())
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI admin API key", err)
		return
	}
	newState := adminAPIKeyResourceModelFromAPI(apiKey, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *adminAPIKeyResource) Update(ctx context.Context, _ resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError("OpenAI admin API keys are immutable", "OpenAI does not expose an update operation for organization admin API keys. Change name or expires_in_seconds by replacing the resource.")
}

func (r *adminAPIKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state adminAPIKeyResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.client.DeleteAdminAPIKey(ctx, state.ID.ValueString()); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI admin API key", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *adminAPIKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func adminAPIKeyResourceModelFromAPI(apiKey *client.AdminAPIKey, prior adminAPIKeyResourceModel) adminAPIKeyResourceModel {
	value := prior.Value
	if apiKey.Value != "" {
		value = types.StringValue(apiKey.Value)
	} else if value.IsUnknown() {
		value = types.StringNull()
	}
	return adminAPIKeyResourceModel{
		ID:               types.StringValue(apiKey.ID),
		Name:             stringOrNull(apiKey.Name),
		ExpiresInSeconds: prior.ExpiresInSeconds,
		Value:            value,
		RedactedValue:    stringOrNull(apiKey.RedactedValue),
		OwnerType:        stringOrNull(apiKey.OwnerType),
		OwnerID:          stringOrNull(apiKey.OwnerID),
		OwnerName:        stringOrNull(apiKey.OwnerName),
		OwnerRole:        stringOrNull(apiKey.OwnerRole),
		CreatedAt:        int64OrNull(apiKey.CreatedAt),
		ExpiresAt:        int64OrNull(apiKey.ExpiresAt),
		LastUsedAt:       int64OrNull(apiKey.LastUsedAt),
	}
}
