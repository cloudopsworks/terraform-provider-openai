package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &adminAPIKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &adminAPIKeyDataSource{}
)

type adminAPIKeyDataSource struct {
	client client.AdminClient
}

type adminAPIKeyDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	RedactedValue types.String `tfsdk:"redacted_value"`
	OwnerType     types.String `tfsdk:"owner_type"`
	OwnerID       types.String `tfsdk:"owner_id"`
	OwnerName     types.String `tfsdk:"owner_name"`
	OwnerRole     types.String `tfsdk:"owner_role"`
	CreatedAt     types.Int64  `tfsdk:"created_at"`
	ExpiresAt     types.Int64  `tfsdk:"expires_at"`
	LastUsedAt    types.Int64  `tfsdk:"last_used_at"`
}

func NewAdminAPIKeyDataSource() datasource.DataSource { return &adminAPIKeyDataSource{} }

func (d *adminAPIKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_api_key"
}

func (d *adminAPIKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Retrieves one OpenAI organization admin API key by ID. The unredacted key value is not returned by the Admin API.",
		Attributes: map[string]datasourceschema.Attribute{
			"id":             datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI admin API key ID."},
			"name":           datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Admin API key name."},
			"redacted_value": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Redacted admin API key value."},
			"owner_type":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner type returned by OpenAI."},
			"owner_id":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner ID returned by OpenAI."},
			"owner_name":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner display name returned by OpenAI."},
			"owner_role":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner role returned by OpenAI."},
			"created_at":     datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for key creation."},
			"expires_at":     datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for key expiration when configured."},
			"last_used_at":   datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the key was last used, if known."},
		},
	}
}

func (d *adminAPIKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	data, err := configureClient(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("OpenAI provider not configured", err.Error())
		return
	}
	d.client = data.client
}

func (d *adminAPIKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config adminAPIKeyDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	apiKey, err := d.client.GetAdminAPIKey(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI admin API key", err)
		return
	}
	state := adminAPIKeyDataSourceModel{
		ID:            types.StringValue(apiKey.ID),
		Name:          stringOrNull(apiKey.Name),
		RedactedValue: stringOrNull(apiKey.RedactedValue),
		OwnerType:     stringOrNull(apiKey.OwnerType),
		OwnerID:       stringOrNull(apiKey.OwnerID),
		OwnerName:     stringOrNull(apiKey.OwnerName),
		OwnerRole:     stringOrNull(apiKey.OwnerRole),
		CreatedAt:     int64OrNull(apiKey.CreatedAt),
		ExpiresAt:     int64OrNull(apiKey.ExpiresAt),
		LastUsedAt:    int64OrNull(apiKey.LastUsedAt),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
