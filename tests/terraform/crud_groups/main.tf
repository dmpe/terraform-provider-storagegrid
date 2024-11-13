# Copyright (c) HashiCorp, Inc.

terraform {
  required_providers {
    storagegrid = {
      source = "github.com/dmpe/storagegrid"
    }
  }
}

provider "storagegrid" {
  address  = "https://grid.firm.com:9443/api/v3"
  username = var.grid_username
  password = var.grid_password
  tenant   = var.grid_tenant_iid
  insecure = true
}

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
  }
}
