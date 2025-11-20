resource "storagegrid_bucket" "example_default_region" {
  name   = "example-bucket-default-region"
}

resource "storagegrid_bucket_versioning" "example" {
  bucket_name = storagegrid_bucket.example_default_region.name

  status = "Enabled"
}
