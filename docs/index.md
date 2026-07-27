# OpenAI provider

This provider uses the OpenAI Administration APIs to manage organization projects,
project service accounts, service-account project API keys, organization admin API
keys, organization groups, group membership, custom organization/project roles,
and user/group role assignments. It uses an existing admin API key supplied
directly or resolved from AWS Secrets Manager, GCP Secret Manager, or Azure Key
Vault.

The current OpenAI Go SDK exposes organization users as read/list/update/delete
surfaces but does not expose user creation. The provider therefore exposes users
as data sources and manages user access through group membership and role
assignment resources rather than creating users directly.
