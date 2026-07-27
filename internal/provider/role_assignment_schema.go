package provider

import (
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func roleAssignmentResourceAttributes(principalName, principalDescription string) map[string]resourceschema.Attribute {
	return map[string]resourceschema.Attribute{
		"id":              resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Composite role assignment ID."},
		principalName:     resourceschema.StringAttribute{Required: true, MarkdownDescription: principalDescription, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"role_id":         resourceschema.StringAttribute{Required: true, MarkdownDescription: "OpenAI role ID to assign.", PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}},
		"name":            resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Assigned role name."},
		"description":     resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Assigned role description."},
		"permissions":     resourceschema.SetAttribute{Computed: true, ElementType: types.StringType, MarkdownDescription: "Permissions granted by the assigned role."},
		"predefined_role": resourceschema.BoolAttribute{Computed: true, MarkdownDescription: "Whether OpenAI predefines and manages the role."},
		"resource_type":   resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Resource type the assigned role applies to."},
		"created_at":      resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the role was created, when returned by OpenAI."},
		"updated_at":      resourceschema.Int64Attribute{Computed: true, MarkdownDescription: "Unix timestamp when the role was last updated, when returned by OpenAI."},
		"created_by":      resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Actor ID that created the role, when returned by OpenAI."},
		"assignment_sources": resourceschema.ListNestedAttribute{Computed: true, MarkdownDescription: "Principals from which the role assignment is inherited, when available.", NestedObject: resourceschema.NestedAttributeObject{Attributes: map[string]resourceschema.Attribute{
			"principal_id":   resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Source principal ID."},
			"principal_type": resourceschema.StringAttribute{Computed: true, MarkdownDescription: "Source principal type."},
		}}},
	}
}
