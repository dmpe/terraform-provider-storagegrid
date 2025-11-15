# Copyright (c) github.com/dmpe
# SPDX-License-Identifier: MIT

variable "vault_domain" {
  description = "Vault domain"
  type        = string
  default     = "test"
}

variable "vault_token" {
  description = "Vault token"
  type        = string
}

variable "grid_tenant_iid" {
  description = "Tenant ID"
  type        = string
}
