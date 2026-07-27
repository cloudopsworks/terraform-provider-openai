package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ datasource.DataSource              = &organizationCertificateDataSource{}
	_ datasource.DataSourceWithConfigure = &organizationCertificateDataSource{}
)

type organizationCertificateDataSource struct{ client client.AdminClient }

type organizationCertificateDataSourceModel struct {
	ID                 types.String            `tfsdk:"id"`
	IncludeContent     types.Bool              `tfsdk:"include_content"`
	Name               types.String            `tfsdk:"name"`
	Object             types.String            `tfsdk:"object"`
	Active             types.Bool              `tfsdk:"active"`
	CertificateDetails certificateDetailsModel `tfsdk:"certificate_details"`
	CreatedAt          types.Int64             `tfsdk:"created_at"`
}

func NewOrganizationCertificateDataSource() datasource.DataSource {
	return &organizationCertificateDataSource{}
}

func (d *organizationCertificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_certificate"
}

func (d *organizationCertificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := organizationCertificateDataSourceAttributes(true)
	attrs["include_content"] = datasourceschema.BoolAttribute{Optional: true, MarkdownDescription: "When true, request certificate_details.content from OpenAI and store it as Sensitive data-source state."}
	resp.Schema = datasourceschema.Schema{MarkdownDescription: "Retrieves one OpenAI organization certificate by ID.", Attributes: attrs}
}

func organizationCertificateDataSourceAttributes(idRequired bool) map[string]datasourceschema.Attribute {
	idAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI certificate ID."}
	if idRequired {
		idAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI certificate ID."}
	}
	return map[string]datasourceschema.Attribute{
		"id":         idAttr,
		"name":       datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Certificate name."},
		"object":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI certificate object type."},
		"active":     datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether the certificate is active at the organization scope."},
		"created_at": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the certificate was uploaded."},
		"certificate_details": datasourceschema.SingleNestedAttribute{Computed: true, MarkdownDescription: "Certificate validity metadata returned by OpenAI.", Attributes: map[string]datasourceschema.Attribute{
			"content":    datasourceschema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Certificate content in PEM format when include_content is true."},
			"expires_at": datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the certificate expires, when returned."},
			"valid_at":   datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the certificate becomes valid, when returned."},
		}},
	}
}

func (d *organizationCertificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *organizationCertificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config organizationCertificateDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}
	includeContent := boolValue(config.IncludeContent)
	certificate, err := d.client.GetOrganizationCertificate(ctx, config.ID.ValueString(), includeContent)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization certificate", err)
		return
	}
	state := organizationCertificateDataSourceModelFromAPI(certificate, includeContent)
	state.IncludeContent = config.IncludeContent
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func organizationCertificateDataSourceModelFromAPI(certificate *client.Certificate, includeContent bool) organizationCertificateDataSourceModel {
	return organizationCertificateDataSourceModel{
		ID:                 types.StringValue(certificate.ID),
		Name:               stringOrNull(certificate.Name),
		Object:             stringOrNull(certificate.Object),
		Active:             types.BoolValue(certificate.Active),
		CertificateDetails: certificateDetailsModelFromAPI(certificate.CertificateDetails, includeContent),
		CreatedAt:          int64OrNull(certificate.CreatedAt),
	}
}
