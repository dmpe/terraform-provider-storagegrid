resource "storagegrid_bucket" "example" {
  name   = "example-bucket"
  region = "example-region"
}

# will use the region configured as default in the StorageGRID instance.
resource "storagegrid_bucket" "example_default_region" {
  name   = "example-bucket-default-region"
}
