data "storagegrid_bucket_policy" "example" {
  bucket_name = "example-bucket-name"
}

output "example_bucket_policy" {
  value = data.storagegrid_bucket_policy.example.policy
}
