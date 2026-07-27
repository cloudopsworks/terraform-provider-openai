package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationUserDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationUserDataSource{}
)

type organizationUserDataSource struct{ client client.AdminClient }

type organizationUserDataSourceModel struct {
	ID                             types.String                   `tfsdk:"id"`
	Name                           types.String                   `tfsdk:"name"`
	Email                          types.String                   `tfsdk:"email"`
	Role                           types.String                   `tfsdk:"role"`
	UserID                         types.String                   `tfsdk:"user_id"`
	UserName                       types.String                   `tfsdk:"user_name"`
	UserEmail                      types.String                   `tfsdk:"user_email"`
	Picture                        types.String                   `tfsdk:"picture"`
	DeveloperPersona               types.String                   `tfsdk:"developer_persona"`
	TechnicalLevel                 types.String                   `tfsdk:"technical_level"`
	AddedAt                        types.Int64                    `tfsdk:"added_at"`
	Created                        types.Int64                    `tfsdk:"created"`
	APIKeyLastUsedAt               types.Int64                    `tfsdk:"api_key_last_used_at"`
	BannedAt                       types.Int64                    `tfsdk:"banned_at"`
	IsDefault                      types.Bool                     `tfsdk:"is_default"`
	IsScaleTierAuthorizedPurchaser types.Bool                     `tfsdk:"is_scale_tier_authorized_purchaser"`
	IsScimManaged                  types.Bool                     `tfsdk:"is_scim_managed"`
	IsServiceAccount               types.Bool                     `tfsdk:"is_service_account"`
	Banned                         types.Bool                     `tfsdk:"banned"`
	Enabled                        types.Bool                     `tfsdk:"enabled"`
	Projects                       []organizationUserProjectModel `tfsdk:"projects"`
}

type organizationUserProjectModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Role types.String `tfsdk:"role"`
}

func NewOrganizationUserDataSource() datasource.DataSource { return &organizationUserDataSource{} }

func (d *organizationUserDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_user"
}

func (d *organizationUserDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI organization user by ID. User creation is not exposed by the current OpenAI Go SDK; manage user access with role and group-membership resources.", Attributes: organizationUserDataSourceAttributes(true)}
}

func organizationUserDataSourceAttributes(idRequired bool) map[string]datasourceschema.Attribute {
	idAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI organization user ID."}
	if idRequired {
		idAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI organization user ID."}
	}
	return map[string]datasourceschema.Attribute{
		"id":                                 idAttr,
		"name":                               datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User display name."},
		"email":                              datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User email address."},
		"role":                               datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Organization role such as owner or reader."},
		"user_id":                            datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Nested user object ID."},
		"user_name":                          datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Nested user object display name."},
		"user_email":                         datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Nested user object email."},
		"picture":                            datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User profile picture URL when available."},
		"developer_persona":                  datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Developer persona metadata when available."},
		"technical_level":                    datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Technical level metadata when available."},
		"added_at":                           datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the user was added."},
		"created":                            datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the user was created."},
		"api_key_last_used_at":               datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for last API key usage when known."},
		"banned_at":                          datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the nested user was banned, if applicable."},
		"is_default":                         datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether this is the organization's default user."},
		"is_scale_tier_authorized_purchaser": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the user is authorized purchaser for Scale Tier."},
		"is_scim_managed":                    datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the user is managed through SCIM."},
		"is_service_account":                 datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the user represents a service account."},
		"banned":                             datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the nested user is banned."},
		"enabled":                            datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the nested user is enabled."},
		"projects": datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Projects associated with the user when returned by OpenAI.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{
			"id":   datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project ID."},
			"name": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project name."},
			"role": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "User role for the project."},
		}}},
	}
}

func (d *organizationUserDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	user, err := d.client.GetOrganizationUser(ctx, config.ID.ValueString())
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization user", err)
		return
	}
	state := organizationUserDataSourceModelFromAPI(user)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func organizationUserDataSourceModelFromAPI(user *client.OrganizationUser) organizationUserDataSourceModel {
	projects := make([]organizationUserProjectModel, 0, len(user.Projects))
	for _, project := range user.Projects {
		projects = append(projects, organizationUserProjectModel{ID: stringOrNull(project.ID), Name: stringOrNull(project.Name), Role: stringOrNull(project.Role)})
	}
	return organizationUserDataSourceModel{
		ID:                             types.StringValue(user.ID),
		Name:                           stringOrNull(user.Name),
		Email:                          stringOrNull(user.Email),
		Role:                           stringOrNull(user.Role),
		UserID:                         stringOrNull(user.UserID),
		UserName:                       stringOrNull(user.UserName),
		UserEmail:                      stringOrNull(user.UserEmail),
		Picture:                        stringOrNull(user.Picture),
		DeveloperPersona:               stringOrNull(user.DeveloperPersona),
		TechnicalLevel:                 stringOrNull(user.TechnicalLevel),
		AddedAt:                        int64OrNull(user.AddedAt),
		Created:                        int64OrNull(user.Created),
		APIKeyLastUsedAt:               int64OrNull(user.APIKeyLastUsedAt),
		BannedAt:                       int64OrNull(user.BannedAt),
		IsDefault:                      types.BoolValue(user.IsDefault),
		IsScaleTierAuthorizedPurchaser: types.BoolValue(user.IsScaleTierAuthorizedPurchaser),
		IsScimManaged:                  types.BoolValue(user.IsScimManaged),
		IsServiceAccount:               types.BoolValue(user.IsServiceAccount),
		Banned:                         types.BoolValue(user.Banned),
		Enabled:                        types.BoolValue(user.Enabled),
		Projects:                       projects,
	}
}
