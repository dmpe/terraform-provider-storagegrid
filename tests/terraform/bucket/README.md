# How To

The Terraform code in this folder can be used to test bucket functionality locally, e.g., using a local build of the
provider.

Please refer to the
[official documentation](https://developer.hashicorp.com/terraform/tutorials/providers-plugin-framework/providers-plugin-framework-provider#prepare-terraform-for-local-provider-install)
of Terraform on how to set up a local development environment.

Please refer to the [variables.tf](variables.tf) file for the required information to input to test the setup.

Specifically, we require buckets to already exist in your StorageGrid account, such that data resources and imports
work as expected with the given configuration.
You can specify bucket names and regions as needed for those buckets.

Most of the resources and functionality provided in this test folder provide some additional tests and verification
for functionality that is not otherwise covered by automatic tests, such as imports of buckets and related resources.

Further explanation can be found in the [main.tf](main.tf) file itself as preamble comments of the affected resources.
