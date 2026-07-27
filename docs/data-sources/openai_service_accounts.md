# openai_service_accounts

API group: [Projects](../api-groups/projects.md).

OpenAI API hierarchy: `Administration > Organization > Projects > Service Accounts`.

Lists OpenAI project service accounts through the Administration API.

## Example Usage

```terraform
data "openai_service_accounts" "all" {
  project_id = "proj_abc123"
  limit      = 100
}
```

## Schema

### Required

- `project_id` (String) OpenAI project ID that owns the service accounts.

### Optional

- `after` (String) Cursor for pagination. Use `last_id` from a previous response to fetch the next page.
- `limit` (Number) Maximum number of service accounts to return. The OpenAI API supports values from 1 through 100.

### Read-Only

- `items` (List of Object) Normalized service account collection returned by the API.
- `has_more` (Boolean) Whether more service accounts are available after this page.
- `last_id` (String) ID of the last service account in the returned page, for use as the `after` cursor.

### `items` Object

- `id` (String) OpenAI service account ID.
- `name` (String) Service account name.
- `role` (String) Project role for the service account.
- `created_at` (Number) Unix timestamp for service-account creation.
