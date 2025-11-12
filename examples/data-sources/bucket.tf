data "storagegrid_bucket" "example" {
  name = "example-bucket-name"
}

output "example_bucket_name" {
  value = data.storagegrid_bucket.example.name
}

output "example_bucket_region" {
  value = data.storagegrid_bucket.example.region
}
