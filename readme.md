# Terraform Provider for NetApp StorageGrid S3

This is a terraform provider plugin for [NetApp StorageGrid S3](https://www.netapp.com/data-storage/storagegrid/) system.

It has been tested & validated to work against [11.7 version](https://docs.netapp.com/us-en/storagegrid-117/).

# Very important information

:warning: There exist 2 git repositories. One is hosted internally and another (this) one is public.

There are many differences.

Among them, the most important given the access to the real-life NetApp StorageGrid, 
are Terraform **tests**. These make sure that this Terraform plugin has been tested and works actually.

Unfortunately, there are far from ideal state and cannot be published completely - at this very moment - in this repository.

For instance, they have dependancies to internal Vault system, and have other issues such as hardcoded URLs, etc.

At such, only a selected number of them is published, without a guarantee that they (will) work when executing them.

# How does sync between internal and this public repo work?

There is no automatic sync'ing available. Any updates are first reviewed and only then pushed to GitHub manually.

# What is working and what is not working?

This provider aims to cover selected **Tenant** [REST API endpoints such](https://docs.netapp.com/us-en/storagegrid/tenant/understanding-tenant-management-api.html) `users`, `groups` or `s3` (which creates access/secret keys). 

This provider does not currently support any [Grid Management API endpoints](https://docs.netapp.com/us-en/storagegrid/admin/grid-management-api-operations.html) which can be found in the Grid Management view.

This severally limits what can be changed and adjusted, when compared to the Grid Management REST API.

# Getting started

Configuring [required providers](https://www.terraform.io/docs/language/providers/requirements.html#requiring-providers):

```terraform
terraform {
  required_providers {
    storagegrid = {
      source  = "dmpe/storagegrid"
      version = "" # My strong advice - always pin this provider's version!
    }
  }
}
```


### Authentication

The StorageGrid provider offers 2 different ways of providing credentials for authentication.

The following methods are supported:

* Static credentials
* Environment variables


#### Static credentials

Default static credentials can be provided by adding the `tenant`, `username`, 
`password` and `address` in the provider block:

Only `insecure` is optional (default is `false`). It could be used when using self-signed certificates on your StorageGrid system.

```terraform
provider "storagegrid" {
  address   = "https://grid.firm.com:9443/api/v3"
  username  = "grid"
  password  = "change_me"
  tenant    = "<int>" # Tenant ID
  insecure  = false
}
```

#### Environment Variables

You can provide your credentials for the default connection via the `STORAGEGRID_ADDRESS`, `STORAGEGRID_USERNAME`, `STORAGEGRID_PASSWORD`, `STORAGEGRID_TENANT` environmental variables. 

Make sure that you export them properly, like this:

```bash
export STORAGEGRID_ADDRESS=
export STORAGEGRID_USERNAME=
export STORAGEGRID_PASSWORD=
export STORAGEGRID_TENANT=
```

```terraform
provider "storagegrid" {
}
```

# Developer Contributions and Documentation

Contributions are always welcome. 

Workflow is simple: 
1. Fork it
2. Push your changes to some branch, and create Pull Request. 
3. Then either ping me or assign me for review.

Please, make sure that your changes either:

- a) include tests (in `golang`, `terraform` or `terraform-in-golang`, etc.) OR
- b) your confirmation that if you cannot publish your tests, your changes have been tested with real StorageGrid system.


## Some additional information: 

- I followed this guideline on creating new plugin: <https://developer.hashicorp.com/terraform/plugin/code-generation/workflow-example>
- We use only the modern [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework). Not SDKv2.


## Code repository structure

- Resource and data sources are located in `internal/provider/`.
- Generated documentation is in `docs/`.
- All tests are in `tests` folder. Golang tests are currently not being developed, i.e. validation is done with Terraform examples.
- `tools` contains some additional functionality such as adding file headers (copywrite) or code for generating documentation. Additionally, in `tools/rest-api`, it contains Swagger/OpenAPI export for specific StorageGrid version. 
- In `root` we find:
  - `makefile` which essentially governs developing this provider. Execute as 

  ```
  make install_dnf
  make lint
  make fmt
  ```
  - `.terraformRC` file which is used for local development.

