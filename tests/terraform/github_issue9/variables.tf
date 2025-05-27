# Copyright (c) github.com/dmpe
# SPDX-License-Identifier: MIT

variable "grid_username" {
  description = "Grid username"
  type        = string
}

variable "grid_password" {
  description = "Grid password"
  type        = string
}

variable "grid_tenant_iid" {
  description = "Tenant ID"
  type        = string
}

variable "groups" {
  description = "List of groups to create"
  type = list(object({
    unique_name  = string
    display_name = string
    management_read_only = optional(bool, true)
    management_policies = object({
      manage_all_containers        = bool
      manage_endpoints             = bool
      manage_own_container_objects = bool
      manage_own_s3_credentials    = bool
      root_access                  = bool
      view_all_containers          = bool
    })
    s3 = object({
        statement = list(object({
          sid      = string
          effect   = string
          action   = list(string)
          resource = list(string)
        }))
    })
  }))
}

variable "users" {
  description = "List of users to create"
  type = list(object({
    unique_name = string
    full_name   = string
    disable     = optional(bool, false)
    member_of   = list(string) # List of group unique_names to assign the user to
    create_key  = optional(bool, false)  # New field to determine if a key should be created
    key_expiry  = optional(string, "") # Optional expiration for the key (ISO 8601 format, e.g., "2028-01-01T00:00:00.000Z")
  }))
}
