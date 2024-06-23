module "gcs_buckets" {
  source  = "terraform-google-modules/cloud-storage/google"
  version = "6.0"

  project_id  = "first"
  names = ["first"]
  prefix = "prefix"
}
