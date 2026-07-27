package provider

import (
	"context"
	"sort"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

func setStringValue(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	return types.SetValueFrom(ctx, types.StringType, sorted)
}

func setToSortedStringSlice(ctx context.Context, value types.Set) ([]string, diag.Diagnostics) {
	values, diags := setToStringSlice(ctx, value)
	if values == nil {
		return nil, diags
	}
	sort.Strings(values)
	return values, diags
}

type roleItemModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	Permissions    types.Set    `tfsdk:"permissions"`
	PredefinedRole types.Bool   `tfsdk:"predefined_role"`
	ResourceType   types.String `tfsdk:"resource_type"`
}

func roleItemModelFromAPI(ctx context.Context, role client.Role) (roleItemModel, diag.Diagnostics) {
	permissions, diags := setStringValue(ctx, role.Permissions)
	return roleItemModel{
		ID:             types.StringValue(role.ID),
		Name:           stringOrNull(role.Name),
		Description:    stringOrNull(role.Description),
		Permissions:    permissions,
		PredefinedRole: types.BoolValue(role.PredefinedRole),
		ResourceType:   stringOrNull(role.ResourceType),
	}, diags
}

type assignmentSourceModel struct {
	PrincipalID   types.String `tfsdk:"principal_id"`
	PrincipalType types.String `tfsdk:"principal_type"`
}

type roleAssignmentItemModel struct {
	ID                types.String            `tfsdk:"id"`
	Name              types.String            `tfsdk:"name"`
	Description       types.String            `tfsdk:"description"`
	Permissions       types.Set               `tfsdk:"permissions"`
	PredefinedRole    types.Bool              `tfsdk:"predefined_role"`
	ResourceType      types.String            `tfsdk:"resource_type"`
	PrincipalID       types.String            `tfsdk:"principal_id"`
	PrincipalType     types.String            `tfsdk:"principal_type"`
	CreatedAt         types.Int64             `tfsdk:"created_at"`
	UpdatedAt         types.Int64             `tfsdk:"updated_at"`
	CreatedBy         types.String            `tfsdk:"created_by"`
	AssignmentSources []assignmentSourceModel `tfsdk:"assignment_sources"`
}

func roleAssignmentItemModelFromAPI(ctx context.Context, assignment client.RoleAssignment) (roleAssignmentItemModel, diag.Diagnostics) {
	permissions, diags := setStringValue(ctx, assignment.Permissions)
	sources := make([]assignmentSourceModel, 0, len(assignment.AssignmentSources))
	for _, source := range assignment.AssignmentSources {
		sources = append(sources, assignmentSourceModel{
			PrincipalID:   stringOrNull(source.PrincipalID),
			PrincipalType: stringOrNull(source.PrincipalType),
		})
	}
	return roleAssignmentItemModel{
		ID:                types.StringValue(assignment.ID),
		Name:              stringOrNull(assignment.Name),
		Description:       stringOrNull(assignment.Description),
		Permissions:       permissions,
		PredefinedRole:    types.BoolValue(assignment.PredefinedRole),
		ResourceType:      stringOrNull(assignment.ResourceType),
		PrincipalID:       stringOrNull(assignment.PrincipalID),
		PrincipalType:     stringOrNull(assignment.PrincipalType),
		CreatedAt:         int64OrNull(assignment.CreatedAt),
		UpdatedAt:         int64OrNull(assignment.UpdatedAt),
		CreatedBy:         stringOrNull(assignment.CreatedBy),
		AssignmentSources: sources,
	}, diags
}
