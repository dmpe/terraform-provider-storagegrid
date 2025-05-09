---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "storagegrid_users Resource - storagegrid"
description: |-
  Create a new user - a resource
---

# storagegrid_users (Resource)

Create a new user - a resource. Must be a member of some group.
See [StorageGrid documentation](https://docs.netapp.com/us-en/storagegrid-118/tenant/managing-local-users.html).

```terraform
resource "storagegrid_users" "new-local-user" {
  unique_name = "user/my_new_test_user_tf_stroragegrid_provider"
  full_name   = "My StorageGrid TF Provider plugin"
  disable     = "false"
  member_of = [
    "a9dd4848-a863-4716-82eb-d0939a6d643b"
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `full_name` (String) The human-readable name for the User (required for local Users and imported automatically for federated Users)
- `member_of` (List of String) Group memberships for this User (required for local Users and imported automatically for federated Users)
- `unique_name` (String) The name this user will use to sign in. Usernames must be unique and cannot be changed.

### Optional

- `disable` (Boolean) Do you want to prevent this user from signing in regardless of assigned group permissions?

### Read-Only

- `account_id` (String)
- `federated` (Boolean) True if the User is federated, for example, an LDAP User
- `id` (String) The ID of this resource.
- `user_urn` (String)

