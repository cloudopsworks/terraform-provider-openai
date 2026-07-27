package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/cloudopsworks/terraform-provider-openai/internal/client"
)

type serviceAccountAPIKeyCreateState struct {
	ID        types.String
	Name      types.String
	Value     types.String
	CreatedAt types.Int64
}

func createProjectAPIKey(ctx context.Context, adminClient client.AdminClient, projectID, serviceAccountID, name string, scopes []string) (*client.ServiceAccountAPIKeyCreateResponse, error) {
	return adminClient.CreateServiceAccountAPIKey(ctx, projectID, serviceAccountID, client.ServiceAccountAPIKeyCreateRequest{
		Name:   name,
		Scopes: append([]string(nil), scopes...),
	})
}

func serviceAccountAPIKeyCreateStateFromAPI(apiKey *client.ServiceAccountAPIKeyCreateResponse) serviceAccountAPIKeyCreateState {
	return serviceAccountAPIKeyCreateState{
		ID:        types.StringValue(apiKey.ID),
		Name:      stringOrNull(apiKey.Name),
		Value:     types.StringValue(apiKey.Value),
		CreatedAt: int64OrNull(apiKey.CreatedAt),
	}
}
