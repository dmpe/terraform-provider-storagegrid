resource "storagegrid_groups" "new-local-group" {
  unique_name          = "group/my_new_test_group_tf_stroragegrid_provider"
  display_name         = "StorageGrid TF Provider plugin"
  management_read_only = "false"
  policies = {
    management = {
      manage_all_containers        = false
      manage_endpoints             = false
      manage_own_container_objects = false
      manage_own_s3_credentials    = true
      root_access                  = false
    }
    s3 = {
      statement = [
        {
          sid      = "681616"
          effect   = "Deny"
          action   = ["s3:GetObject"]
          resource = ["arn:aws:s3:::mybucket/myobject"]
        }
      ]
    }
    s3_trust = {
      version = "2012-10-17"
      statement = [
        {
          effect        = "Allow"
          action        = ["sts:AssumeRole"]
          principal_aws = ["urn:sgws:identity::12345678901234567890:user/vault"]
        }
      ]
    }
  }
}
