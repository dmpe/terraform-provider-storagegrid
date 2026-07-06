resource "storagegrid_bucket" "example_default_region" {
  name = "example-bucket-default-region"
}

data "storagegrid_user" "root_user" {
  unique_name = "root"
}

variable "grid_tenant_iid" {
  type        = string
  sensitive   = true
  description = "ID of the Storage Grid tenant."
}

resource "storagegrid_bucket_policy" "example" {
  bucket_name = storagegrid_bucket.example_default_region.name

  policy = {
    statement = [{
      sid      = "example-sid"
      effect   = "Allow"
      action   = ["s3:ListBucket"]
      resource = ["arn:aws:s3:::${storagegrid_bucket.example_default_region.name}", "arn:aws:s3:::${storagegrid_bucket.example_default_region.name}/*"]
      principal = {
        type        = "AWS"
        identifiers = ["arn:aws:iam::${var.grid_tenant_iid}:${data.storagegrid_user.root_user.unique_name}"]
      }
    }]
  }
}
