data "openai_projects" "all" {
  limit            = 100
  include_archived = true
}

