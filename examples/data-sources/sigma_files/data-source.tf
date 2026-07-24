data "sigma_files" "example" {
  parent_id            = "folder-id"
  direct_children_only = true
  type_filters         = ["folder", "workbook", "report"]
}
