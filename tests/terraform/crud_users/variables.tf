# Copyright (c) github.com/dmpe
# SPDX-License-Identifier: MIT

variable "grid_tenant_iid" {
  type        = string
  description = "Tenant ID"
}

variable "grid_username" {
  type        = string
  description = "User name"
}

variable "grid_password" {
  type        = string
  description = "Password"
}

variable "grid_url" {
  type        = string
  description = "Grid URL"
}

variable "group_memberships" {
  type        = list(string)
  description = "Groups the test user is supposed to be a member of."
  default     = []
}
