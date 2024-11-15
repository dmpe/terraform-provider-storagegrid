---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "storagegrid_group Data Source - storagegrid"
subcategory: ""
description: |-
  Fetch a specific group - a data source
---

# storagegrid_group (Data Source)

Fetch a specific group - a data source



<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `unique_name` (String)

### Read-Only

- `account_id` (String)
- `display_name` (String)
- `federated` (Boolean)
- `group_urn` (String)
- `id` (String) The ID of this resource.
- `management_read_only` (Boolean)
- `policies` (Attributes) (see [below for nested schema](#nestedatt--policies))

<a id="nestedatt--policies"></a>
### Nested Schema for `policies`

Read-Only:

- `management` (Attributes) (see [below for nested schema](#nestedatt--policies--management))
- `s3` (Attributes) (see [below for nested schema](#nestedatt--policies--s3))

<a id="nestedatt--policies--management"></a>
### Nested Schema for `policies.management`

Read-Only:

- `manage_all_containers` (Boolean)
- `manage_endpoints` (Boolean)
- `manage_own_container_objects` (Boolean)
- `manage_own_s3_credentials` (Boolean)
- `root_access` (Boolean)


<a id="nestedatt--policies--s3"></a>
### Nested Schema for `policies.s3`

Required:

- `statement` (Attributes List) a list of group policy statements (see [below for nested schema](#nestedatt--policies--s3--statement))

Optional:

- `id` (String) S3 Policy ID provided by policy generator tools (currently not used)
- `version` (String) S3 API Version (currently not used)

<a id="nestedatt--policies--s3--statement"></a>
### Nested Schema for `policies.s3.statement`

Optional:

- `action` (List of String) the specific actions that will be allowed (Can be a string if only one element. A statement must have either Action or NotAction.)
- `effect` (String) the specific result of the statement (either an allow or an explicit deny)
- `not_action` (List of String) the specific exceptional actions (Can be a string if only one element. A statement must have either Action or NotAction.)
- `not_resource` (List of String) the objects that the statement does not cover (Can be a string if only one element. A statement must have either Resource or NotResource.)
- `resource` (List of String) the objects that the statement covers (Can be a string if only one element. A statement must have either Resource or NotResource.)
- `sid` (String) an optional identifier that you provide for the policy statement
