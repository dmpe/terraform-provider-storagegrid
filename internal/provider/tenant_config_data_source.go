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
								Description: "The maximum retention period in days allowed for new objects in compliance or governance mode. Does not affect existing objects. If both maxRetentionDays and maxRetentionYears are null, the maximum retention limit will be 100 years.",
							},
							"max_retention_years": schema.Int64Attribute{
								Computed:    true,
								Description: "The maximum retention period in years allowed for new objects in compliance or governance mode. Does not affect existing objects. If both maxRetentionDays and maxRetentionYears are null, the maximum retention limit will be 100 years.",
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
				Attributes: map[string]schema.Attribute{},
			},
			"token": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
			},
			"permissions": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
			},
			"deactivated_features": schema.SingleNestedBlock{
				Attributes: map[string]schema.Attribute{},
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
		// ...
	}

	type tenantConfigModelToken struct {
		// ...
	}

	type tenantConfigModelPermissions struct {
		// ...
	}

	type tenantConfigModelFeatures struct {
		// ...
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
		DeactivatedFeatures tenantConfigModelFeatures    `json:"deactivated_features"`
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
		// user        TenantConfigModelUser
		// token       TenantConfigModelToken
		// permissions TenantConfigModelPermissions
		// features    TenantConfigModelFeatures
		account TenantConfigModelAccount
		policy  TenantConfigModelAccountPolicy
		data    TenantConfigModel
	)

	// Translate API model to Terraform model
	account.ID = types.StringValue(config.Data.Account.ID)
	account.Name = types.StringValue(config.Data.Account.Name)
	account.Capabilities = make([]types.String, len(config.Data.Account.Capabilities))
	for i, cap := range config.Data.Account.Capabilities {
		account.Capabilities[i] = types.StringValue(cap)
	}
	policy.UseAccountIdentitySource = types.BoolValue(config.Data.Account.Policy.UseAccountIdentitySource)
	policy.AllowPlatformServices = types.BoolValue(config.Data.Account.Policy.AllowPlatformServices)
	policy.AllowSelectObjectContent = types.BoolValue(config.Data.Account.Policy.AllowSelectObjectContent)
	policy.AllowComplianceMode = types.BoolValue(config.Data.Account.Policy.AllowComplianceMode)
	policy.MaxRetentionDays = types.Int64Value(int64(config.Data.Account.Policy.MaxRetentionDays))
	policy.MaxRetentionYears = types.Int64Value(int64(config.Data.Account.Policy.MaxRetentionYears))
	policy.QuotaObjectBytes = types.Int64Value(int64(config.Data.Account.Policy.QuotaObjectBytes))
	policy.AllowedGridFederationConnections = types.StringValue(config.Data.Account.Policy.AllowedGridFederationConnections)
	account.Policy = &policy
	account.Replica = types.BoolValue(config.Data.Account.Replica)

	data.AutoLogout = types.Int64Value(int64(config.Data.AutoLogout))
	// data.User = &user
	// data.Token = &token
	// data.Permissions = &permissions
	// data.DeactivatedFeatures = &features
	data.Account = &account
	data.RestrictedPort = types.BoolValue(config.Data.RestrictedPort)

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
