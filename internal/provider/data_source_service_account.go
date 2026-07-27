package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &serviceAccountDataSource{}
	_ datasource.DataSourceWithConfigure = &serviceAccountDataSource{}
)

type serviceAccountDataSource struct {
	client client.AdminClient
}

type serviceAccountDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	ProjectID types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Role      types.String `tfsdk:"role"`
	CreatedAt types.Int64  `tfsdk:"created_at"`
}

func NewServiceAccountDataSource() datasource.DataSource { return &serviceAccountDataSource{} }

func (d *serviceAccountDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *serviceAccountDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Looks up an OpenAI project service account by ID or exact name through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"id":         datasourceschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "OpenAI service account ID. When set, lookup uses the service account retrieve endpoint."},
			"project_id": datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI project ID that owns the service account."},
			"name":       datasourceschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Exact service account name. When id is omitted, lookup scans the project's service account list for this name."},
			"role":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Project role for the service account."},
			"created_at": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp for service-account creation."},
		},
	}
}

func (d *serviceAccountDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *serviceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config serviceAccountDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	projectID := stringValue(config.ProjectID)
	id := stringValue(config.ID)
	name := stringValue(config.Name)
	if id == "" && name == "" {
		resp.Diagnostics.AddError("Missing service account lookup value", "Set either id or name to look up an OpenAI service account.")
		return
	}

	var account *client.ServiceAccount
	var err error
	if id != "" {
		account, err = d.client.GetServiceAccount(ctx, projectID, id)
		if err == nil && name != "" && account.Name != name {
			resp.Diagnostics.AddError("OpenAI service account name mismatch", fmt.Sprintf("Service account %q in project %q has name %q, not %q.", id, projectID, account.Name, name))
			return
		}
	} else {
		account, err = d.lookupServiceAccountByName(ctx, projectID, name)
	}
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI service account", err)
		return
	}

	state := serviceAccountDataSourceModel{
		ID:        types.StringValue(account.ID),
		ProjectID: config.ProjectID,
		Name:      stringOrNull(account.Name),
		Role:      stringOrNull(account.Role),
		CreatedAt: int64OrNull(account.CreatedAt),
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (d *serviceAccountDataSource) lookupServiceAccountByName(ctx context.Context, projectID, name string) (*client.ServiceAccount, error) {
	var after string
	var matched *client.ServiceAccount
	for {
		page, err := d.client.ListServiceAccounts(ctx, projectID, client.ServiceAccountListRequest{After: after, Limit: 100})
		if err != nil {
			return nil, err
		}
		for i := range page.Items {
			account := page.Items[i]
			if account.Name != name {
				continue
			}
			if matched != nil {
				return nil, fmt.Errorf("found multiple OpenAI service accounts named %q in project %q", name, projectID)
			}
			matched = &account
		}
		if !page.HasMore || page.LastID == "" {
			break
		}
		after = page.LastID
	}
	if matched == nil {
		return nil, fmt.Errorf("no OpenAI service account named %q was found in project %q", name, projectID)
	}
	return matched, nil
}
