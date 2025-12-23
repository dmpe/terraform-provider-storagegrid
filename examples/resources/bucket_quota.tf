resource "storagegrid_bucket" "example_default_region" {
  name   = "example-bucket-default-region"
}

resource "storagegrid_bucket_quota" "example" {
  bucket_name = storagegrid_bucket.example_default_region.name

  object_bytes = 10000000
}
