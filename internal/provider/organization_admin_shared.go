package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

const organizationSingletonID = "organization"

var (
	organizationDataRetentionTypes = []string{
		"zero_data_retention",
		"modified_abuse_monitoring",
		"enhanced_zero_data_retention",
		"enhanced_modified_abuse_monitoring",
	}
	inviteRoles        = []string{"reader", "owner"}
	projectMemberRoles = []string{"member", "owner"}
	spendCurrencies    = []string{"USD"}
	spendIntervals     = []string{"month"}
)

type stringEnumValidator struct {
	values []string
}

func newStringEnumValidator(values ...string) stringEnumValidator {
	return stringEnumValidator{values: append([]string(nil), values...)}
}

func (v stringEnumValidator) Description(_ context.Context) string {
	return fmt.Sprintf("value must be one of: %s", strings.Join(v.values, ", "))
}

func (v stringEnumValidator) MarkdownDescription(ctx context.Context) string {
	return v.Description(ctx)
}

func (v stringEnumValidator) ValidateString(_ context.Context, req validator.StringRequest, resp *validator.StringResponse) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	value := req.ConfigValue.ValueString()
	for _, allowed := range v.values {
		if value == allowed {
			return
		}
	}
	resp.Diagnostics.AddAttributeError(req.Path, "Invalid OpenAI Administration API value", v.Description(context.Background())+".")
}

type inviteProjectModel struct {
	ID   types.String `tfsdk:"id"`
	Role types.String `tfsdk:"role"`
}

func inviteProjectsFromModel(projects []inviteProjectModel) []client.InviteProject {
	if projects == nil {
		return nil
	}
	result := make([]client.InviteProject, 0, len(projects))
	for _, project := range projects {
		result = append(result, client.InviteProject{ID: stringValue(project.ID), Role: stringValue(project.Role)})
	}
	return result
}

func inviteProjectModelsFromAPI(projects []client.InviteProject) []inviteProjectModel {
	if projects == nil {
		return nil
	}
	result := make([]inviteProjectModel, 0, len(projects))
	for _, project := range projects {
		result = append(result, inviteProjectModel{ID: stringOrNull(project.ID), Role: stringOrNull(project.Role)})
	}
	return result
}

type spendAlertNotificationChannelModel struct {
	Recipients    types.Set    `tfsdk:"recipients"`
	Type          types.String `tfsdk:"type"`
	SubjectPrefix types.String `tfsdk:"subject_prefix"`
}

func spendAlertNotificationChannelFromModel(ctx context.Context, model spendAlertNotificationChannelModel, attrPath path.Path) (client.SpendAlertNotificationChannel, diag.Diagnostics) {
	var diags diag.Diagnostics
	recipients, recipientDiags := setToSortedStringSlice(ctx, model.Recipients)
	diags.Append(recipientDiags...)
	for i, recipient := range recipients {
		if strings.TrimSpace(recipient) == "" {
			diags.AddAttributeError(attrPath.AtName("recipients").AtSetValue(types.StringValue(recipient)), "Invalid OpenAI spend alert recipient", "Spend alert recipient email addresses cannot be empty.")
			continue
		}
		recipients[i] = strings.TrimSpace(recipient)
	}
	if len(recipients) == 0 {
		diags.AddAttributeError(attrPath.AtName("recipients"), "Invalid OpenAI spend alert recipients", "At least one recipient email address is required.")
	}
	channelType := stringValue(model.Type)
	if channelType == "" {
		channelType = "email"
	}
	return client.SpendAlertNotificationChannel{
		Recipients:    recipients,
		Type:          channelType,
		SubjectPrefix: stringValue(model.SubjectPrefix),
	}, diags
}

func spendAlertNotificationChannelModelFromAPI(ctx context.Context, channel client.SpendAlertNotificationChannel) (spendAlertNotificationChannelModel, diag.Diagnostics) {
	recipients := append([]string(nil), channel.Recipients...)
	sort.Strings(recipients)
	recipientSet, diags := setStringValue(ctx, recipients)
	channelType := channel.Type
	if channelType == "" {
		channelType = "email"
	}
	return spendAlertNotificationChannelModel{
		Recipients:    recipientSet,
		Type:          stringOrNull(channelType),
		SubjectPrefix: stringOrNull(channel.SubjectPrefix),
	}, diags
}

type certificateDetailsModel struct {
	Content   types.String `tfsdk:"content"`
	ExpiresAt types.Int64  `tfsdk:"expires_at"`
	ValidAt   types.Int64  `tfsdk:"valid_at"`
}

type certificateItemModel struct {
	ID                 types.String            `tfsdk:"id"`
	Name               types.String            `tfsdk:"name"`
	Object             types.String            `tfsdk:"object"`
	Active             types.Bool              `tfsdk:"active"`
	CertificateDetails certificateDetailsModel `tfsdk:"certificate_details"`
	CreatedAt          types.Int64             `tfsdk:"created_at"`
}

func certificateDetailsModelFromAPI(details client.CertificateDetails, includeContent bool) certificateDetailsModel {
	content := types.StringNull()
	if includeContent && strings.TrimSpace(details.Content) != "" {
		content = types.StringValue(details.Content)
	}
	return certificateDetailsModel{
		Content:   content,
		ExpiresAt: int64OrNull(details.ExpiresAt),
		ValidAt:   int64OrNull(details.ValidAt),
	}
}

func certificateItemModelFromAPI(certificate client.Certificate, includeContent bool) certificateItemModel {
	return certificateItemModel{
		ID:                 types.StringValue(certificate.ID),
		Name:               stringOrNull(certificate.Name),
		Object:             stringOrNull(certificate.Object),
		Active:             types.BoolValue(certificate.Active),
		CertificateDetails: certificateDetailsModelFromAPI(certificate.CertificateDetails, includeContent),
		CreatedAt:          int64OrNull(certificate.CreatedAt),
	}
}
