// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"os"
)

// Ensure storagegridProvider satisfies various provider interfaces.
var (
	_ provider.Provider              = &storagegridProvider{}
	_ provider.ProviderWithFunctions = &storagegridProvider{}
)

// storagegridProvider defines the provider implementation.
type storagegridProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// storagegridProviderModel describes the provider data model.
type storagegridProviderModel struct {
	Address            types.String `tfsdk:"address"`
	Username           types.String `tfsdk:"username"`
	Password           types.String `tfsdk:"password"`
	Tenant             types.String `tfsdk:"tenant"`
	EnableTraceContext types.Bool   `tfsdk:"enable_trace_context"`
	Insecure           types.Bool   `tfsdk:"insecure"`
}

func (p *storagegridProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "storagegrid"
	resp.Version = p.version
}

func (p *storagegridProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"address": schema.StringAttribute{
				Description: "The address of StorageGrid tenant. Full FQDN with port number if some non-standard is used. Must be without '/' at the end.",
				Optional:    false,
				Required:    true,
				Sensitive:   false,
			},
			"username": schema.StringAttribute{
				Description: "Provider username.",
				Optional:    true,
				Required:    true,
				Sensitive:   false,
			},
			"password": schema.StringAttribute{
				Description: "Provider password.",
				Optional:    true,
				Required:    true,
				Sensitive:   true,
			},
			"tenant": schema.StringAttribute{
				Description: "Provide tenant id.",
				Optional:    true,
				Required:    true,
				Sensitive:   false,
			},
			"enable_trace_context": schema.BoolAttribute{
				Optional:            true,
				MarkdownDescription: "Enable trace context. If `true` a `Traceparent` header will be added to the request. Default: `false`",
			},
			"insecure": schema.BoolAttribute{
				Optional:            true,
				Description:         "Use insecure HTTP connection",
				MarkdownDescription: "Use insecure HTTP connection? Setting this to true will ignore certificates. Default: `false`",
			},
		},
	}
}

func (p *storagegridProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data storagegridProviderModel
	var insecure bool
	tflog.Debug(ctx, "Configuring StorageGrid client.")

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if data.Address.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("address"),
			"Unknown StorageGrid Address",
			"The provider must have StorageGrid address specified. "+
				"Either set address value statically in the configuration, or use the STORAGEGRID_ADDRESS environment variable.",
		)
	}

	if data.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown StorageGrid username",
			"The provider must have StorageGrid username specified. "+
				"Either set address value statically in the configuration, or use the STORAGEGRID_USERNAME environment variable.",
		)
	}

	if data.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown StorageGrid password",
			"The provider must have StorageGrid password specified. "+
				"Either set address value statically in the configuration, or use the STORAGEGRID_PASSWORD environment variable.",
		)
	}

	if data.Tenant.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant"),
			"Unknown StorageGrid tenant",
			"The provider must have StorageGrid tenant specified. "+
				"Either set address value statically in the configuration, or use the STORAGEGRID_TENANT environment variable.",
		)
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.
	address := os.Getenv("STORAGEGRID_ADDRESS")
	username := os.Getenv("STORAGEGRID_USERNAME")
	password := os.Getenv("STORAGEGRID_PASSWORD")
	tenant := os.Getenv("STORAGEGRID_TENANT")
	trc_ctxt := os.Getenv("TF_ACC")

	if !data.Address.IsNull() {
		address = data.Address.ValueString()
	}

	if !data.Username.IsNull() {
		username = data.Username.ValueString()
	}

	if !data.Password.IsNull() {
		password = data.Password.ValueString()
	}

	if !data.Tenant.IsNull() {
		tenant = data.Tenant.ValueString()
	}

	if !data.Insecure.IsNull() {
		insecure = data.Insecure.ValueBool()
	}

	if trc_ctxt == "1" {
		data.EnableTraceContext = types.BoolValue(true)
	}

	if insecure {
		data.Insecure = types.BoolValue(true)
	}

	if address == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("address"),
			"Missing StorageGrid address",
			"The provider cannot create the StorageGrid API client as there is a missing or empty value for the StorageGrid API host. "+
				"Set the address value in the configuration or use the STORAGEGRID_ADDRESS environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing StorageGrid username",
			"The provider cannot create the StorageGrid API client as there is a missing or empty value for the StorageGrid API username. "+
				"Set the username value in the configuration or use the STORAGEGRID_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing StorageGrid password",
			"The provider cannot create the StorageGrid API client as there is a missing or empty value for the StorageGrid API password. "+
				"Set the password value in the configuration or use the STORAGEGRID_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if tenant == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("tenant"),
			"Missing StorageGrid tenant",
			"The provider cannot create the StorageGrid API client as there is a missing or empty value for the StorageGrid API tenant. "+
				"Set the tenant value in the configuration or use the STORAGEGRID_TENANT environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	ctx = tflog.SetField(ctx, "storagegrid_address", address)
	ctx = tflog.SetField(ctx, "storagegrid_username", username)
	ctx = tflog.SetField(ctx, "storagegrid_password", password)
	ctx = tflog.SetField(ctx, "storagegrid_tenant", tenant)

	clientUsPsw := NewUsernamePasswordClient(
		address,
		username,
		password,
		tenant,
		insecure,
	)
	authZToken, _, _ := clientUsPsw.SendAuthorizeRequest(200)
	client := NewTokenClient(address, authZToken, insecure)
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Debug(ctx, "Configuration of StorageGrid client is finished.")
}

func (p *storagegridProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewGroupsResource,
		NewUsersResource,
		NewS3AccessSecretKeyResource,
	}
}

func (p *storagegridProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewGroupsDataSource,
		NewGroupDataSource,
		NewUsersDataSource,
		NewUserDataSource,
		NewS3DataSource_ByUserID_All,
		NewS3DataSource_ByUserID_AccountID,
	}
}

func (p *storagegridProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &storagegridProvider{
			version: version,
		}
	}
}
