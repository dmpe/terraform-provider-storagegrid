# Terraform Provider for NetApp StorageGRID S3

This is a terraform provider plugin for [NetApp StorageGRID S3](https://www.netapp.com/data-storage/storagegrid/) system.

**Update Summer 2025**:
I no longer work at the company which has access to the StorageGrid system. 
As a result, I am no longer able to conduct any tests and rely on contributors to 1) find and 2) fix any issues. 

That does __NOT__ mean that this provider is abandoned and will not be updated anymore. 
If fact, if a PR is opened, I will be happy to review + release a new version ASAP.

## What is working and what is not working?

This provider aims to cover selected **Tenant** [REST API endpoints such](https://docs.netapp.com/us-en/storagegrid/tenant/understanding-tenant-management-api.html) `users`, `groups`, `buckets` or `s3` (which creates access/secret keys). 

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

`insecure` can be used when using self-signed certificates on your StorageGRID system.

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

You can also provide your credentials for the default connection via the `STORAGEGRID_ADDRESS`, 
`STORAGEGRID_USERNAME`, `STORAGEGRID_PASSWORD`, `STORAGEGRID_TENANT` environmental variables. 

Make sure that you export them properly, like this:

```bash
export STORAGEGRID_ADDRESS=
export STORAGEGRID_USERNAME=
export STORAGEGRID_PASSWORD=
export STORAGEGRID_TENANT=
```

and then use:

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

## Terraform Acceptance Tests

To run the Terraform acceptance tests, set the following environment variables:

- `STORAGEGRID_TEST_GRID_VERSION` to the version of your StorageGRID instance
- `STORAGEGRID_TEST_FEDERATED_USER` to the username of a federated user that exists in your StorageGRID instance
  - only for StorageGRID versions 12.0 and higher
- `STORAGEGRID_TEST_DEFAULT_REGION` to the default region of your StorageGRID instance

Subsequently, execute the `make testacc` command.

> [!WARNING] Real resources are created
> Executing Terraform acceptance tests create real resources in your StorageGRID instance.
> Those resources are also destroyed again as part of the test procedure.
> Generally, names for resources created by the acceptance tests are prefixed with `terraform-acc-test-` and contain a
> random string, however, naming conflicts might occur nonetheless.

> [!NOTE] Test resources are required
> Acceptance Tests of the provider's data sources require the read resources to already exist in the StorageGRID instance
> used for testing.
> Make sure to create those resources before running the acceptance tests.

### Required Test Resources

To run the Terraform acceptance tests, the following resources and configurations must be present in the StorageGRID
instance used for testing.

- **buckets**
  - `tf-provider-acc-test-bucket`
    - located in the region `us-east-1`
    - having a capacity limit of 1 GB
    - having the S3 object lock disabled
    - and the following bucket policy attached
```json
{
  "Statement": [
    {
      "Sid": "test-sid-1",
      "Effect": "Allow",
      "Action": "s3:ListBucket",
      "Resource": [
        "arn:aws:s3:::tf-provider-acc-test-bucket",
        "arn:aws:s3:::tf-provider-acc-test-bucket/*"
      ],
      "Principal": "*"
    },
    {
      "Sid": "test-sid-2",
      "Effect": "Allow",
      "Action": [
        "s3:ListBucket",
        "s3:GetObject"
      ],
      "Resource": [
        "arn:aws:s3:::tf-provider-acc-test-bucket",
        "arn:aws:s3:::tf-provider-acc-test-bucket/*"
      ],
      "Principal": "*"
    }
  ]
}
```
  - `tf-provider-acc-test-bucket-ol`
    - located in the region `us-east-1`
    - having the S3 object lock enabled with mode `GOVERNANCE` and a retention period of 10 days
- **federated users** (only StorageGRID version 12 or newer)
  - an identity federation set up in which a user with a given name can be imported
    - use the `STORAGEGRID_TEST_FEDERATED_USER` environment variable to specify the name of the user to import

## Contributors

This project has received significant contributions by:

- several corporate employees employed at <https://github.com/svalabs>.

Other open source enthusiasts can be found here:
- <https://github.com/dmpe/terraform-provider-storagegrid/graphs/contributors>

I am grateful to them. Thank you.
