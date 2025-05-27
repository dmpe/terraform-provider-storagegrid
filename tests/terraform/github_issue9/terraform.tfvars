# Copyright (c) github.com/dmpe
# SPDX-License-Identifier: MIT

groups = [
  {
    unique_name  = "group/demo"
    display_name = "Demo"
    management_policies = {
      manage_all_containers        = false
      manage_endpoints             = false
      manage_own_container_objects = true
      manage_own_s3_credentials    = true
      root_access                  = false
      view_all_containers          = false
    }
    s3 = {
      statement = [
        {
          sid      = "deny-policy"
          effect   = "Deny"
          action   = ["s3:*"]
          resource = ["arn:aws:s3:::"]
        }
      ]
    }
  }
]

users = [
  {
    unique_name = "user/bill"
    full_name   = "Bill"
    disable     = false
    member_of   = ["group/demo"]
    create_key  = true
    key_expiry  = "2026-01-01T00:00:00.000Z"
  },
  {
    unique_name = "user/jill"
    full_name   = "Jill"
    disable     = false
    member_of   = ["group/demo"]
    create_key  = false
  },

]
