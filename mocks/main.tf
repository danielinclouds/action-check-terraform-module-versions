variable "project_id" {
  type = string
  default = "main"
}

module "gcs_buckets" {
  source  = "terraform-google-modules/cloud-storage/google"
  version = "6.0"

  project_id  = var.project_id
  names = ["main"]
  prefix = "prefix"
}

module "gcs_buckets2" {
  source  = "terraform-google-modules/cloud-storage/google"
  version = "5.0"

  project_id  = var.project_id
  names = ["main2"]
  prefix = "prefix"
}

module "gcs_buckets_private" {
  source  = "git::https://github.com/danielinclouds/terraform-module-example.git?ref=0.0.1"

  project_id  = var.project_id
  names = ["main2"]
  prefix = "prefix"
}
