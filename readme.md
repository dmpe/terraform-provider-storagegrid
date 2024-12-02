# Terraform Provider for NetApp StorageGRID S3

This is a terraform provider plugin for [NetApp StorageGRID S3](https://www.netapp.com/data-storage/storagegrid/) system.

Version `v1.0.0` has been tested & validated to work against [11.8 version](https://docs.netapp.com/us-en/storagegrid-118/).

## What is working and what is not working?

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

The StorageGRID provider offers 2 different ways of providing credentials for authentication.

The following methods are supported:

* Static credentials
* Environment variables

#### Static credentials

Default static credentials can be provided by adding the `tenant`, `username`, 
`password` and `address` in the provider block:

Only `insecure` is optional (default is `false`). It could be used when using self-signed certificates on your StorageGRID system.

```terraform
provider "storagegrid" {
  address   = "https://grid.firm.com:9443"
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

## How to develop this provider

Contributions are always welcome! In order to develop this provider your system needs:

- `make`
- `golang`
- `terraform` for running real life tests

The GitHub workflow is very simple:

1. Fork this repo.
2. Push your changes to some branch, and create Pull Request against this repo.
3. Then either ping me or assign me for review.

Please, make sure that your changes either:

- a) include tests (in `golang`, `terraform` or `terraform-in-golang`, etc.) OR
- b) your confirmation that if you cannot publish your tests, your changes have been tested with real StorageGRID system.

## Some additional information:

- I followed this guideline fow how to create new provider: <https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework>.
- We use only the modern [Terraform Plugin Framework](https://github.com/hashicorp/terraform-plugin-framework) SDK.
Not the `SDKv2` and support for it (whatever the reason) will not be accepted.
- An early attempt was done with [Code Generation](https://developer.hashicorp.com/terraform/plugin/code-generation) approach, but I
have failed to overcome several issues with StorageGRID REST API (=json file) without doing manual changes to the Swagger API.


## Code repository structure

- `Resources` and `data sources` are located in `internal/provider/`.
- Generated documentation is in `docs/`. Trigger it by `make generate`.
- All tests are in `tests` folder. Golang tests are currently not being developed, i.e. validation is done with real life Terraform examples.
- `tools` folder contains some additional functionality such as adding file headers (`copywrite`) or code for aforementioned generation of documentation. 
Additionally, in `tools/rest-api`, it contains Swagger/OpenAPI export for specific StorageGRID version(s). 
- In `root` we can find:
  - `.terraformrc` file which is used for local development. You may not need it. But you will, if your tests will include other Terraform providers (such as my internal tests that use HashiCorp Vault etc.)
  - `makefile` which essentially governs developing this provider. Execute as 

  ```bash
  make install_dnf
  make lint
  make fmt
  make build
  ```

The repo also contains `Dockerfile` which can be build using `make docker`. 
After that you simply use inside the container different `make` commands like this:

```bash
docker run -it -v $(pwd):/home storagegrid_dev:latest
$ make build
$ make lint
....
```
