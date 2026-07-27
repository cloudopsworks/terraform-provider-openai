package provider

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

func configureClient(data any) (*providerData, error) {
	providerData, ok := data.(*providerData)
	if !ok || providerData == nil || providerData.client == nil {
		return nil, fmt.Errorf("provider not configured")
	}
	return providerData, nil
}

func stringOrNull(value string) types.String {
	if strings.TrimSpace(value) == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func int64OrNull(value int64) types.Int64 {
	if value == 0 {
		return types.Int64Null()
	}
	return types.Int64Value(value)
}

func int64Value(value types.Int64) int64 {
	if value.IsNull() || value.IsUnknown() {
		return 0
	}
	return value.ValueInt64()
}

func boolValue(value types.Bool) bool {
	if value.IsNull() || value.IsUnknown() {
		return false
	}
	return value.ValueBool()
}

func setStringValueOrNull(ctx context.Context, values []string) (types.Set, diag.Diagnostics) {
	if len(values) == 0 {
		return types.SetNull(types.StringType), nil
	}
	sorted := append([]string(nil), values...)
	sort.Strings(sorted)
	return types.SetValueFrom(ctx, types.StringType, sorted)
}

func setToStringSlice(ctx context.Context, value types.Set) ([]string, diag.Diagnostics) {
	if value.IsNull() || value.IsUnknown() {
		return nil, nil
	}
	var values []string
	diags := value.ElementsAs(ctx, &values, false)
	sort.Strings(values)
	return values, diags
}

func setOrNull(value types.Set) types.Set {
	if value.IsUnknown() {
		return types.SetNull(types.StringType)
	}
	return value
}

func stringOrNullFromState(value types.String) types.String {
	if value.IsUnknown() {
		return types.StringNull()
	}
	return value
}

func int64OrNullFromState(value types.Int64) types.Int64 {
	if value.IsUnknown() {
		return types.Int64Null()
	}
	return value
}

func parseTwoPartImportID(id, first, second string) (string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", fmt.Errorf("expected import format <%s>/<%s>", first, second)
	}
	return parts[0], parts[1], nil
}

func parseThreePartImportID(id, first, second, third string) (string, string, string, error) {
	parts := strings.Split(id, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", fmt.Errorf("expected import format <%s>/<%s>/<%s>", first, second, third)
	}
	return parts[0], parts[1], parts[2], nil
}

func addClientError(diags *diag.Diagnostics, summary string, err error) {
	diags.AddError(summary, client.ErrorSummary(err))
}
