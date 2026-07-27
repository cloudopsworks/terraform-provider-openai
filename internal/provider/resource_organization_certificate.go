package provider

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

var (
	_ resource.Resource                = &organizationCertificateResource{}
	_ resource.ResourceWithConfigure   = &organizationCertificateResource{}
	_ resource.ResourceWithImportState = &organizationCertificateResource{}
)

type organizationCertificateResource struct{ client client.AdminClient }

type organizationCertificateResourceModel struct {
	ID                 types.String            `tfsdk:"id"`
	Name               types.String            `tfsdk:"name"`
	Certificate        types.String            `tfsdk:"certificate"`
	Active             types.Bool              `tfsdk:"active"`
	Object             types.String            `tfsdk:"object"`
	CertificateDetails certificateDetailsModel `tfsdk:"certificate_details"`
	CreatedAt          types.Int64             `tfsdk:"created_at"`
}

func NewOrganizationCertificateResource() resource.Resource {
	return &organizationCertificateResource{}
}

func (r *organizationCertificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization_certificate"
}

func (r *organizationCertificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = resourceschema.Schema{
		MarkdownDescription: "OpenAI organization certificate. Uploading a certificate does not automatically activate it; set active = true to activate it at the organization scope after upload.",
		Attributes: map[string]resourceschema.Attribute{
			"id":          resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI certificate ID."},
			"name":        resourceschema.StringAttribute{Optional: true, Computed: true, MarkdownDescription: "Certificate name. OpenAI supports updating only the name after upload."},
			"certificate": resourceschema.StringAttribute{Optional: true, Sensitive: true, MarkdownDescription: "Certificate content in PEM format. Required for create; not returned on normal refresh and changes replace the certificate.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
			"active":      resourceschema.BoolAttribute{Optional: true, Computed: true, Default: booldefault.StaticBool(false), MarkdownDescription: "Whether the certificate should be active at the organization scope."},
			"object":      resourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI certificate object type."},
			"created_at":  resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the certificate was uploaded."},
			"certificate_details": resourceschema.SingleNestedAttribute{Computed: true, MarkdownDescription: "Certificate validity metadata returned by OpenAI.", Attributes: map[string]resourceschema.Attribute{
				"content":    resourceschema.StringAttribute{Computed: true, Sensitive: true, MarkdownDescription: "Certificate content when explicitly fetched by an API path that returns it. The resource normally keeps this null and preserves the create input in certificate."},
				"expires_at": resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the certificate expires, when returned."},
				"valid_at":   resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the certificate becomes valid, when returned."},
			}},
		},
	}
}

func (r *organizationCertificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *organizationCertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan organizationCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if strings.TrimSpace(stringValue(plan.Certificate)) == "" {
		resp.Diagnostics.AddAttributeError(path.Root("certificate"), "Missing OpenAI organization certificate content", "certificate is required when creating an organization certificate resource.")
		return
	}
	certificate, err := r.client.CreateOrganizationCertificate(ctx, client.CertificateCreateRequest{Name: stringValue(plan.Name), Certificate: plan.Certificate.ValueString()})
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to create OpenAI organization certificate", err)
		return
	}
	if boolValue(plan.Active) {
		if _, err := r.client.SetOrganizationCertificatesActive(ctx, []string{certificate.ID}, true); err != nil {
			if cleanupErr := r.client.DeleteOrganizationCertificate(ctx, certificate.ID); cleanupErr != nil && !client.IsNotFound(cleanupErr) {
				resp.Diagnostics.AddWarning("OpenAI organization certificate cleanup failed", client.ErrorSummary(cleanupErr))
			}
			addClientError(&resp.Diagnostics, "Unable to activate OpenAI organization certificate", err)
			return
		}
		certificate, err = r.client.GetOrganizationCertificate(ctx, certificate.ID, false)
		if err != nil {
			addClientError(&resp.Diagnostics, "Unable to read activated OpenAI organization certificate", err)
			return
		}
	}
	state := organizationCertificateResourceModelFromAPI(certificate, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *organizationCertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state organizationCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	certificate, err := r.client.GetOrganizationCertificate(ctx, state.ID.ValueString(), false)
	if err != nil {
		if client.IsNotFound(err) {
			resp.State.RemoveResource(ctx)
			return
		}
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization certificate", err)
		return
	}
	newState := organizationCertificateResourceModelFromAPI(certificate, state)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationCertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan organizationCertificateResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	var state organizationCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := plan.ID.ValueString()
	if !plan.Name.IsNull() && !plan.Name.IsUnknown() && plan.Name.ValueString() != stringValue(state.Name) {
		if _, err := r.client.UpdateOrganizationCertificate(ctx, id, client.CertificateUpdateRequest{Name: plan.Name.ValueString()}); err != nil {
			addClientError(&resp.Diagnostics, "Unable to update OpenAI organization certificate", err)
			return
		}
	}
	if boolValue(plan.Active) != boolValue(state.Active) {
		if _, err := r.client.SetOrganizationCertificatesActive(ctx, []string{id}, boolValue(plan.Active)); err != nil {
			addClientError(&resp.Diagnostics, "Unable to update OpenAI organization certificate activation", err)
			return
		}
	}
	certificate, err := r.client.GetOrganizationCertificate(ctx, id, false)
	if err != nil {
		addClientError(&resp.Diagnostics, "Unable to read OpenAI organization certificate", err)
		return
	}
	newState := organizationCertificateResourceModelFromAPI(certificate, plan)
	resp.Diagnostics.Append(resp.State.Set(ctx, &newState)...)
}

func (r *organizationCertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state organizationCertificateResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.ID.ValueString()
	if boolValue(state.Active) {
		if _, err := r.client.SetOrganizationCertificatesActive(ctx, []string{id}, false); err != nil && !client.IsNotFound(err) {
			addClientError(&resp.Diagnostics, "Unable to deactivate OpenAI organization certificate", err)
			return
		}
	}
	if err := r.client.DeleteOrganizationCertificate(ctx, id); err != nil && !client.IsNotFound(err) {
		addClientError(&resp.Diagnostics, "Unable to delete OpenAI organization certificate", err)
		return
	}
	resp.State.RemoveResource(ctx)
}

func (r *organizationCertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func organizationCertificateResourceModelFromAPI(certificate *client.Certificate, prior organizationCertificateResourceModel) organizationCertificateResourceModel {
	resourceCertificate := prior.Certificate
	if resourceCertificate.IsUnknown() {
		resourceCertificate = types.StringNull()
	}
	return organizationCertificateResourceModel{
		ID:                 types.StringValue(certificate.ID),
		Name:               stringOrNull(certificate.Name),
		Certificate:        resourceCertificate,
		Active:             types.BoolValue(certificate.Active),
		Object:             stringOrNull(certificate.Object),
		CertificateDetails: certificateDetailsModelFromAPI(certificate.CertificateDetails, false),
		CreatedAt:          int64OrNull(certificate.CreatedAt),
	}
}
