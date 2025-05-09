---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "storagegrid_s3_user_id_access_key Data Source - storagegrid"
subcategory: ""
description: |-
  Access specific S3 access key for a user - a data source
---

# storagegrid_s3_user_id_access_key (Data Source)

Access specific S3 access key for a user - a data source



<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `access_key` (String)
- `user_uuid` (String)

### Read-Only

- `data` (Attributes) the response data for the request (required on success and optional on error; type and content vary by request) (see [below for nested schema](#nestedatt--data))

<a id="nestedatt--data"></a>
### Nested Schema for `data`

Read-Only:

- `account_id` (String)
- `display_name` (String)
- `expires` (String)
- `id` (String)
- `user_urn` (String)
- `user_uuid` (String)
