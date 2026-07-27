package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationCertificatesDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationCertificatesDataSource{}
)

type organizationCertificatesDataSource struct{ client client.AdminClient }

type organizationCertificatesDataSourceModel struct {
	After   types.String           `tfsdk:"after"`
	Limit   types.Int64            `tfsdk:"limit"`
	Order   types.String           `tfsdk:"order"`
	Items   []certificateItemModel `tfsdk:"items"`
	HasMore types.Bool             `tfsdk:"has_more"`
	LastID  types.String           `tfsdk:"last_id"`
}

func NewOrganizationCertificatesDataSource() datasource.DataSource {
	return &organizationCertificatesDataSource{}
}

func (d *organizationCertificatesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_certificates"
}

func (d *organizationCertificatesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = datasourceschema.Schema{
		MarkdownDescription: "Lists OpenAI organization certificates through the Administration API.",
		Attributes: map[string]datasourceschema.Attribute{
			"after":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Cursor for pagination."},
			"limit":    datasourceschema.Int64Attribute{Optional: true, MarkdownDescription: "Maximum number of certificates to return."},
			"order":    datasourceschema.StringAttribute{Optional: true, MarkdownDescription: "Sort order by creation time: asc or desc."},
			"items":    datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Organization certificates returned by the API.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: organizationCertificateDataSourceAttributes(false)}},
			"has_more": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether more certificates are available after this page."},
			"last_id":  datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "ID of the last certificate in the page."},
		},
	}
}

func (d *organizationCertificatesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationCertificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationCertificatesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	page, err := d.client.ListOrganizationCertificates(ctx, client.CertificateListRequest{After: stringValue(config.After), Limit: int64Value(config.Limit), Order: stringValue(config.Order)})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to list OpenAI organization certificates", err)
		return
	}
	state := organizationCertificatesDataSourceModel{After: config.After, Limit: config.Limit, Order: config.Order, Items: make([]certificateItemModel, 0, len(page.Items)), HasMore: types.BoolValue(page.HasMore), LastID: stringOrNull(page.LastID)}
	for _, certificate := range page.Items {
		state.Items = append(state.Items, certificateItemModelFromAPI(certificate, false))
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
