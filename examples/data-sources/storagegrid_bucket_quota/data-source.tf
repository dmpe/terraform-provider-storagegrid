data "storagegrid_bucket_quota" "example" {
  bucket_name = "example-bucket-name"
}

output "example_bucket_quota_object_bytes" {
  value = data.storagegrid_bucket_quota.example.object_bytes
}
