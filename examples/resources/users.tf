resource "storagegrid_users" "new-local-user" {
  unique_name = "user/my_new_test_user_tf_stroragegrid_provider"
  full_name   = "My StorageGrid TF Provider plugin"
  disable     = "false"
  member_of = [
    "a9dd4848-a863-4716-82eb-d0939a6d643b"
  ]
}
