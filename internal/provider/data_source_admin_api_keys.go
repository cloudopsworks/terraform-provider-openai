package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &adminAPIKeysDataSource{}
	_ datasource.DataSourceWithConfigure = &adminAPIKeysDataSource{}
)

type adminAPIKeysDataSource struct {
	client client.AdminClient
}

type adminAPIKeysDataSourceModel struct {
	After   types.String                 `tfsdk:"after"`
	Limit   types.Int64                  `tfsdk:"limit"`
	Order   types.String                 `tfsdk:"order"`
	Items   []adminAPIKeyDataSourceModel `tfsdk:"items"`
	HasMore types.Bool                   `tfsdk:"has_more"`
	LastID  types.String                 `tfsdk:"last_id"`
}

func NewAdminAPIKeysDataSource() datasource.DataSource { return &adminAPIKeysDataSource{} }

func (d *adminAPIKeysDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_admin_api_keys"
}

func (d *adminAPIKeysDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI organization admin API keys. Values are redacted by the Admin API.",
		Attributes: map[string]datasourceschema.Attribute{
			"after":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":    datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of keys to return."},
			"order":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order by creation time: asc or desc."},
			"items":    datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Admin API keys returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: adminAPIKeyDataSourceAttributes(false)}},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more keys are available after this page."},
			"last_id":  datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the last key in the returned page, for use as the after cursor."},
		},
	}
}

func adminAPIKeyDataSourceAttributes(idRequired bool) map[string]datasourceschema.Attribute {
	idAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI admin API key ID."}
	if idRequired {
		idAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI admin API key ID."}
	}
	return map[string]datasourceschema.Attribute{
		"id":             idAttr,
		"name":           datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Admin API key name."},
		"redacted_value": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Redacted admin API key value."},
		"owner_type":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner type returned by OpenAI."},
		"owner_id":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner ID returned by OpenAI."},
		"owner_name":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner display name returned by OpenAI."},
		"owner_role":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Owner role returned by OpenAI."},
		"created_at":     datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for key creation."},
		"expires_at":     datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for key expiration when configured."},
		"last_used_at":   datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the key was last used, if known."},
	}
}

func (d *adminAPIKeysDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *adminAPIKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config adminAPIKeysDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListAdminAPIKeys(ctx, client.AdminAPIKeyListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI admin API keys", err)
		return
	}
	state := adminAPIKeysDataSourceModel{After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]adminAPIKeyDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), LastID: stringOrNull(page.LastID)}
	for _, apiKey := range page.Items {
		state.Items = append(state.Items, adminAPIKeyDataSourceModel{
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
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
