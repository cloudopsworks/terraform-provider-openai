# openai_projects

API group: [Projects](../api-groups/projects.md).

OpenAI API hierarchy: `Administration > Organization > Projects`.

Lists OpenAI organization projects through the Administration API.

## Example Usage

```terraform
data "openai_projects" "all" {
  limit            = 100
  include_archived = true
}
```

## Schema

### Optional

- `after` (String) Cursor for pagination. Use `last_id` from a previous response to fetch the next page.
- `limit` (Number) Maximum number of projects to return. The OpenAI API supports values from 1 through 100.
- `include_archived` (Boolean) Include archived projects. Archived projects are excluded by default.

### Read-Only

- `items` (List of Object) Normalized project collection returned by the API.
- `has_more` (Boolean) Whether more projects are available after this page.
- `last_id` (String) ID of the last project in the returned page, for use as the `after` cursor.

Each `items` element includes:

- `id` (String) OpenAI project ID.
- `name` (String) Project display name.
- `external_key_id` (String) External key ID associated with the project, when present.
- `status` (String) Project status, such as `active` or `archived`.
- `created_at` (Number) Unix timestamp for project creation.
- `archived_at` (Number) Unix timestamp for project archival when archived.
