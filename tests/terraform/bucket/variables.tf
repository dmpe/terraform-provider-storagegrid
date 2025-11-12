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

variable "example_bucket" {
  type = object({
    name = string
    region = string
  })
  description = "Name and region of the example buckets to be created for testing."
}

variable "import_bucket" {
  type = object({
    name   = string
    region = string
  })
  description = "Information about existing bucket in StorageGRID. Will be imported for testing."
}

variable "read_bucket_name" {
  type        = string
  description = "Name of existing bucket in StorageGRID. Will be read as data source."
}
