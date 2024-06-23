module "gcs_buckets" {
  source  = "terraform-google-modules/cloud-storage/google"
  version = "6.0"

  project_id  = "second"
  names = ["second"]
  prefix = "prefix"
}
