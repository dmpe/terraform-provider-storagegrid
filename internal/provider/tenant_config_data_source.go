// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &tenantConfigDataSource{}
var _ datasource.DataSourceWithConfigure = &tenantConfigDataSource{}

// NewTenantConfigDataSource returns a new resource instance.
func NewTenantConfigDataSource() datasource.DataSource {
	return &tenantConfigDataSource{}
}

// tenantConfigDataSource defines the data source implementation.
type tenantConfigDataSource struct {
	client *S3GridClient
}

func (d *tenantConfigDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tenant_config"
}

func (d *tenantConfigDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetch the global configuration - a data source",

		Attributes: map[string]schema.Attribute{
			"auto_logout": schema.Int64Attribute{
				Computed:    true,
				Description: "The timeout period for the browser session in seconds (zero for disabled)",
			},
			"restricted_port": schema.BoolAttribute{
				Computed:    true,
				Description: "Whether your current connection is using a restricted port that allows access to Tenant Management APIs (/org) but prevents access to Grid Management APIs (/grid)",
			},
		},
		Blocks: map[string]schema.Block{
			"account": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "A unique identifier for the account (automatically assigned when an account is created)",
					},
					"name": schema.StringAttribute{
						Computed:    true,
						Description: "The descriptive name specified for the account (this name is for display only and might not be unique)",
					},
					"capabilities": schema.SetAttribute{
						ElementType: types.StringType,
						Computed:    true,
						Description: "The high-level features enabled for this account, such as S3 or Swift protocols (accounts must have the 'management' capability if users will sign into the Tenant Manager)",
					},
					"policy": schema.SingleNestedAttribute{
						Computed:    true,
						Description: "Settings for the tenant account",
						Attributes: map[string]schema.Attribute{
							"use_account_identity_source": schema.BoolAttribute{
								Computed:    true,
								Description: "Whether the tenant account should configure its own identity source. If false, the tenant uses the grid-wide identity source.",
							},
							"allow_platform_services": schema.BoolAttribute{
								Computed:    true,
								Description: "Whether a tenant can use platform services features such as CloudMirror. These features send data to an external service that is specified using a StorageGRID endpoint.",
							},
							"allow_select_object_content": schema.BoolAttribute{
								Computed:    true,
								Description: "Whether a tenant can use the S3 SelectObjectContent API to filter and retrieve object data.",
							},
							"allow_compliance_mode": schema.BoolAttribute{
								Computed:    true,
								Description: "Whether a tenant can use compliance mode for object lock and retention.",
							},
							"max_retention_days": schema.Int64Attribute{
								Computed:    true,
								Description: "The maximum retention period in days allowed for new objects in compliance or governance mode. Does not affect existing objects. If both max_retention_days and max_retention_years are null, the maximum retention limit will be 100 years.",
							},
							"max_retention_years": schema.Int64Attribute{
								Computed:    true,
								Description: "The maximum retention period in years allowed for new objects in compliance or governance mode. Does not affect existing objects. If both max_retention_days and max_retention_years are null, the maximum retention limit will be 100 years.",
							},
							"quota_object_bytes": schema.Int64Attribute{
								Computed:    true,
								Description: "The maximum number of bytes available for this tenant's objects. Represents a logical amount (object size), not a physical amount (size on disk). If null, an unlimited number of bytes is available.",
							},
							"allowed_grid_federation_connections": schema.StringAttribute{
								Computed:    true,
								Description: "Connection IDs of tenants. When specified, cross-grid replication of this account and the buckets in this account will be allowed.",
							},
						},
					},
					"replica": schema.BoolAttribute{
						Computed:    true,
						Description: "Automatically assigned when generating the response. Ignored in the PUT body. Present only if this tenant account has permission to use a grid federation connection. If true, this account on the local grid is a replica of an account created on another grid. If false, this account was created on the local grid and is not a copy.",
					},
				},
			},
			"user": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"id": schema.StringAttribute{
						Computed:    true,
						Description: "UUID for the User (generated automatically)",
					},
					"username": schema.StringAttribute{
						Computed:    true,
						Description: "The username that is used to sign in",
					},
					"unique_name": schema.StringAttribute{
						Computed:    true,
						Description: "The machine-readable name for the User (unique within an Account)",
					},
					"first_name": schema.StringAttribute{
						Computed:    true,
						Description: "The User's first name",
					},
					"full_name": schema.StringAttribute{
						Computed:    true,
						Description: "The human-readable name for the User",
					},
					"federated": schema.BoolAttribute{
						Computed:    true,
						Description: "True if the User is federated, for example, an LDAP User",
					},
					"management_read_only": schema.BoolAttribute{
						Computed:    true,
						Description: "Whether the user is in a read-only group",
					},
				},
			},
			"token": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"expires": schema.StringAttribute{
						Computed:    true,
						Description: "Time when the token expires",
					},
				},
			},
			"permissions": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"manage_all_containers": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to manage all S3 buckets or Swift containers for this tenant account (overrides permission settings in group or bucket policies). Supersedes the view_all_containers permission.",
					},
					"manage_endpoints": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to manage all S3 endpoints for this tenant account",
					},
					"manage_own_s3_credentials": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to manage your personal S3 credentials",
					},
					"manage_own_container_objects": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to use S3 Console to view and manage bucket objects",
					},
					"view_all_containers": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to view settings for all S3 buckets or Swift containers for this tenant account. Superseded by the manage_all_containers permission.",
					},
					"root_access": schema.BoolAttribute{
						Computed:    true,
						Description: "Full access to all tenant administration features",
					},
				},
			},
			"deactivated_features": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{
					"manage_all_containers": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to manage all S3 buckets or Swift containers for this tenant account (overrides permission settings in group or bucket policies). Supersedes the view_all_containers permission.",
					},
					"manage_endpoints": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to manage all S3 endpoints for this tenant account",
					},
					"manage_own_s3_credentials": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to manage your personal S3 credentials",
					},
					"manage_own_container_objects": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to use S3 Console to view and manage bucket objects",
					},
					"view_all_containers": schema.BoolAttribute{
						Computed:    true,
						Description: "Ability to view settings for all S3 buckets or Swift containers for this tenant account. Superseded by the manage_all_containers permission.",
					},
				},
			},
		},
	}
}

func (d *tenantConfigDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*S3GridClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *tenantConfigDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state TenantConfigModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}

	respBody, _, _, err := d.client.SendRequest("GET", api_config, nil, 200)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read tenant config, got error: %s", err))
		return
	}

	type tenantConfigModelUser struct {
		ID                 string `json:"id"`
		Username           string `json:"username"`
		UniqueName         string `json:"uniqueName"`
		FirstName          string `json:"firstName"`
		FullName           string `json:"fullName"`
		Federated          bool   `json:"federated"`
		ManagementReadOnly bool   `json:"managementReadOnly"`
	}

	type tenantConfigModelToken struct {
		Expires string `json:"expires"`
	}

	type tenantConfigModelPermissions struct {
		ManageAllContainers       bool `json:"manageAllContainers"`
		ManageEndpoints           bool `json:"manageEndpoints"`
		ManageOwnS3Credentials    bool `json:"manageOwnS3Credentials"`
		ManageOwnContainerObjects bool `json:"manageOwnContainerObjects"`
		ViewAllContainers         bool `json:"viewAllContainers"`
		RootAccess                bool `json:"rootAccess"`
	}

	type tenantConfigModelFeatures struct {
		ManageAllContainers       bool `json:"manageAllContainers"`
		ManageEndpoints           bool `json:"manageEndpoints"`
		ManageOwnS3Credentials    bool `json:"manageOwnS3Credentials"`
		ManageOwnContainerObjects bool `json:"manageOwnContainerObjects"`
		ViewAllContainers         bool `json:"viewAllContainers"`
	}

	type tenantConfigModelAccountPolicy struct {
		UseAccountIdentitySource         bool   `json:"use_account_identity_source"`
		AllowPlatformServices            bool   `json:"allow_platform_services"`
		AllowSelectObjectContent         bool   `json:"allow_select_object_content"`
		AllowComplianceMode              bool   `json:"allow_compliance_mode"`
		MaxRetentionDays                 int    `json:"max_retention_days"`
		MaxRetentionYears                int    `json:"max_retention_years"`
		QuotaObjectBytes                 int    `json:"quota_object_bytes"`
		AllowedGridFederationConnections string `json:"allowed_grid_federation_connections"`
	}

	type tenantConfigModelAccount struct {
		ID           string                          `json:"id"`
		Name         string                          `json:"name"`
		Capabilities []string                        `json:"capabilities"`
		Policy       *tenantConfigModelAccountPolicy `json:"policy"`
		Replica      bool                            `json:"accountReplica"`
	}

	type tenantConfigModel struct {
		AutoLogout          int                          `json:"auto-logout"`
		User                tenantConfigModelUser        `json:"user"`
		Token               tenantConfigModelToken       `json:"token"`
		Permissions         tenantConfigModelPermissions `json:"permissions"`
		DeactivatedFeatures tenantConfigModelFeatures    `json:"deactivatedFeatures"`
		Account             tenantConfigModelAccount     `json:"account"`
		RestrictedPort      bool                         `json:"restrictedPort"`
	}

	type tenantConfigReadModel struct {
		Data tenantConfigModel `json:"data"`
	}

	var config tenantConfigReadModel

	// Map response body to API model
	if err := json.Unmarshal(respBody, &config); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	var (
		user        TenantConfigModelUser
		token       TenantConfigModelToken
		permissions TenantConfigModelPermissions
		features    TenantConfigModelFeatures
		account     TenantConfigModelAccount
		policy      TenantConfigModelAccountPolicy
		data        TenantConfigModel
	)

	// Translate API model to Terraform model
	user.ID = types.StringValue(config.Data.User.ID)
	user.Username = types.StringValue(config.Data.User.Username)
	user.UniqueName = types.StringValue(config.Data.User.UniqueName)
	user.FirstName = types.StringValue(config.Data.User.FirstName)
	user.FullName = types.StringValue(config.Data.User.FullName)
	user.Federated = types.BoolValue(config.Data.User.Federated)
	user.ManagementReadOnly = types.BoolValue(config.Data.User.ManagementReadOnly)

	token.Expires = types.StringValue(config.Data.Token.Expires)

	permissions.ManageAllContainers = types.BoolValue(config.Data.Permissions.ManageAllContainers)
	permissions.ManageEndpoints = types.BoolValue(config.Data.Permissions.ManageEndpoints)
	permissions.ManageOwnS3Credentials = types.BoolValue(config.Data.Permissions.ManageOwnS3Credentials)
	permissions.ManageOwnContainerObjects = types.BoolValue(config.Data.Permissions.ManageOwnContainerObjects)
	permissions.ViewAllContainers = types.BoolValue(config.Data.Permissions.ViewAllContainers)
	permissions.RootAccess = types.BoolValue(config.Data.Permissions.RootAccess)

	features.ManageAllContainers = types.BoolValue(config.Data.DeactivatedFeatures.ManageAllContainers)
	features.ManageEndpoints = types.BoolValue(config.Data.DeactivatedFeatures.ManageEndpoints)
	features.ManageOwnS3Credentials = types.BoolValue(config.Data.DeactivatedFeatures.ManageOwnS3Credentials)
	features.ManageOwnContainerObjects = types.BoolValue(config.Data.DeactivatedFeatures.ManageOwnContainerObjects)
	features.ViewAllContainers = types.BoolValue(config.Data.DeactivatedFeatures.ViewAllContainers)

	policy.UseAccountIdentitySource = types.BoolValue(config.Data.Account.Policy.UseAccountIdentitySource)
	policy.AllowPlatformServices = types.BoolValue(config.Data.Account.Policy.AllowPlatformServices)
	policy.AllowSelectObjectContent = types.BoolValue(config.Data.Account.Policy.AllowSelectObjectContent)
	policy.AllowComplianceMode = types.BoolValue(config.Data.Account.Policy.AllowComplianceMode)
	policy.MaxRetentionDays = types.Int64Value(int64(config.Data.Account.Policy.MaxRetentionDays))
	policy.MaxRetentionYears = types.Int64Value(int64(config.Data.Account.Policy.MaxRetentionYears))
	policy.QuotaObjectBytes = types.Int64Value(int64(config.Data.Account.Policy.QuotaObjectBytes))
	policy.AllowedGridFederationConnections = types.StringValue(config.Data.Account.Policy.AllowedGridFederationConnections)

	account.ID = types.StringValue(config.Data.Account.ID)
	account.Name = types.StringValue(config.Data.Account.Name)
	account.Capabilities = make([]types.String, len(config.Data.Account.Capabilities))
	for i, cap := range config.Data.Account.Capabilities {
		account.Capabilities[i] = types.StringValue(cap)
	}
	account.Policy = &policy
	account.Replica = types.BoolValue(config.Data.Account.Replica)

	data.AutoLogout = types.Int64Value(int64(config.Data.AutoLogout))
	data.User = &user
	data.Token = &token
	data.Permissions = &permissions
	data.DeactivatedFeatures = &features
	data.Account = &account
	data.RestrictedPort = types.BoolValue(config.Data.RestrictedPort)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
