## 1.17.0

- Update dependencies.
- [the S3 secret access keys minted by storagegrid_s3_access_key and storagegrid_s3_access_key_current_user are not marked Sensitive](https://github.com/dmpe/terraform-provider-storagegrid/security/advisories/GHSA-qjqv-jx7x-c7g6) as fixed by @kta1kri. Thanks.


## 1.14.0

- Update of documentation.
- Update dependencies.
- [Fix user read handling without existing ID](https://github.com/dmpe/terraform-provider-storagegrid/pull/46) as fixed by @Lenn97-prog. Thanks.


## 1.12/13.0

- Update of documentation.

## 1.11.0

Bug Fix:
- [User creation with multiple group memberships may lead to inconsistent apply](https://github.com/dmpe/terraform-provider-storagegrid/issues/36) as reported by @cmdaltent. Thanks.


## 1.10.0

Feature:
- Add support for [managing object quota and policies on buckets](https://github.com/dmpe/terraform-provider-storagegrid/pull/34) as reported by @cmdaltent. Thanks.


## 1.9.0

Feature:
- Add support for [Bucket Versioning and Object Locking](https://github.com/dmpe/terraform-provider-storagegrid/pull/31) as reported by @cmdaltent. Thanks.


## 1.8.0

Feature:
- Add support for [buckets](https://github.com/dmpe/terraform-provider-storagegrid/issues/28) as reported by @cmdaltent. Thanks.

Bug fix:
- Update dependencies.

## 1.7.0

Bug Fix:
- Fix [this issue](https://github.com/dmpe/terraform-provider-storagegrid/issues/22) as reported by @mamoep. Thanks.
- Update dependencies.

## 1.6.0

Bug Fix:

- Fix [this issue](https://github.com/dmpe/terraform-provider-storagegrid/pull/19) as reported by @jorijn. Thanks.
- Update documentation a bit.

## 1.5.0

Bug Fix:

- Fix [this issue](https://github.com/dmpe/terraform-provider-storagegrid/pull/18) as reported by @dglauche. Thanks.


## 1.4.0

Bug Fix:

- Fix [this issue](https://github.com/dmpe/terraform-provider-storagegrid/pull/16) as reported by @dglauche. Thanks.


## 1.3.0

Bug Fix:

- An attempt to fix [this issue](https://github.com/dmpe/terraform-provider-storagegrid/issues/9) as reported by @bajo. Thanks.


## 1.2.0

FEATURES:

- Add more examples in the documentation and further expand it

## 1.1.0

FEATURES:

- Expand documentation and update it
- Add new examples in the documentation

## 1.0.0

FEATURES:

- initial release to github and attempting to release the same version to Terraform Provider Registry
