package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationUsersDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationUsersDataSource{}
)

type organizationUsersDataSource struct{ client client.AdminClient }

type organizationUsersDataSourceModel struct {
	After   types.String                      `tfsdk:"after"`
	Limit   types.Int64                       `tfsdk:"limit"`
	Emails  types.Set                         `tfsdk:"emails"`
	Items   []organizationUserDataSourceModel `tfsdk:"items"`
	HasMore types.Bool                        `tfsdk:"has_more"`
	LastID  types.String                      `tfsdk:"last_id"`
}

func NewOrganizationUsersDataSource() datasource.DataSource { return &organizationUsersDataSource{} }

func (d *organizationUsersDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_users"
}

func (d *organizationUsersDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI organization users through the Administration API. Optional emails filter narrows the API response.",
		Attributes: map[string]datasourceschema.Attribute{
			"after":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":    datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of users to return."},
			"emails":   datasourceschema.SetAttribute{Optional: true, ElementType: types.StringType, MarkdownDescription: "Optional set of user email addresses to filter."},
			"items":    datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Organization users returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: organizationUserDataSourceAttributes(false)}},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more users are available after this page."},
			"last_id":  datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the last user in the returned page."},
		},
	}
}

func (d *organizationUsersDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationUsersDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	emails, diags := setToSortedStringSlice(ctx, config.Emails)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	for i, email := range emails {
		if email == "" {
			resp.Diagnostics.AddAttributeError(path.Root("emails").AtListIndex(i), "Invalid OpenAI organization user email filter", "Email filters must not be empty.")
		}
	}
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListOrganizationUsers(ctx, client.OrganizationUserListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Emails: emails})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI organization users", err)
		return
	}
	state := organizationUsersDataSourceModel{After: config.After, Limit: config.Limit, Emails: config.Emails, Items: make([]organizationUserDataSourceModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), LastID: stringOrNull(page.LastID)}
	for _, user := range page.Items {
		state.Items = append(state.Items, organizationUserDataSourceModelFromAPI(&user))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
