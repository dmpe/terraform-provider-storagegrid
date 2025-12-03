data "storagegrid_bucket_versioning" "example" {
  bucket_name = "example-bucket-name"
}

output "example_bucket_versioning_status" {
  value = data.storagegrid_bucket_versioning.example.status
}
