package provider

import (
	datasourceschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func roleAssignmentDataSourceAttributes(principalName, principalDescription string, single bool) map[string]datasourceschema.Attribute {
	principalAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: principalDescription}
	roleIDAttr := datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "OpenAI role ID."}
	if single {
		principalAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: principalDescription}
		roleIDAttr = datasourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI role ID."}
	}
	return map[string]datasourceschema.Attribute{
		"id":              datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Composite role assignment ID."},
		principalName:     principalAttr,
		"role_id":         roleIDAttr,
		"name":            datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Assigned role name."},
		"description":     datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Assigned role description."},
		"permissions":     datasourceschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Permissions granted by the assigned role."},
		"predefined_role": datasourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether OpenAI predefines and manages the role."},
		"resource_type":   datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Resource type the assigned role applies to."},
		"created_at":      datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the role was created, when returned by OpenAI."},
		"updated_at":      datasourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the role was last updated, when returned by OpenAI."},
		"created_by":      datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Actor ID that created the role, when returned by OpenAI."},
		"assignment_sources": datasourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Principals from which the role assignment is inherited, when available.", NestedObject: datasourceschema.NestedAttributeObject{Attributes: map[string]datasourceschema.Attribute{
			"principal_id":   datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Source principal ID."},
			"principal_type": datasourceschema.StringAttribute{Computed: true, MarkdownDescription: "Source principal type."},
		}}},
	}
}
