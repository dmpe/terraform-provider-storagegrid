// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ datasource.DataSource              = &groupsDataSource{}
	_ datasource.DataSourceWithConfigure = &groupsDataSource{}
)

func NewGroupsDataSource() datasource.DataSource {
	return &groupsDataSource{}
}

// groupsDataSource defines the data source implementation.
type groupsDataSource struct {
	client *S3GridClient
}

func (d *groupsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (d *groupsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "groups data source",

		Attributes: map[string]schema.Attribute{
			"data": schema.ListNestedAttribute{
				Computed:            true,
				Description:         "the response data for the request (required on success and optional on error; type and content vary by request)",
				MarkdownDescription: "the response data for the request (required on success and optional on error; type and content vary by request)",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"account_id": schema.StringAttribute{
							Computed: true,
						},
						"display_name": schema.StringAttribute{
							Computed: true,
						},
						"unique_name": schema.StringAttribute{
							Computed: true,
						},
						"group_urn": schema.StringAttribute{
							Computed: true,
						},
						"federated": schema.BoolAttribute{
							Computed: true,
						},
						"management_read_only": schema.BoolAttribute{
							Computed: true,
						},
						"policies": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"management": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"manage_all_containers": schema.BoolAttribute{
											Computed: true,
										},
										"manage_endpoints": schema.BoolAttribute{
											Computed: true,
										},
										"manage_own_container_objects": schema.BoolAttribute{
											Computed: true,
										},
										"manage_own_s3_credentials": schema.BoolAttribute{
											Computed: true,
										},
										"root_access": schema.BoolAttribute{
											Computed: true,
										},
									},
								},
								"s3": schema.SingleNestedAttribute{
									Computed: true,
									Attributes: map[string]schema.Attribute{
										"id": schema.StringAttribute{
											Optional:            true,
											Computed:            true,
											Description:         "S3 Policy ID provided by policy generator tools (currently not used)",
											MarkdownDescription: "S3 Policy ID provided by policy generator tools (currently not used)",
										},
										"statement": schema.ListNestedAttribute{
											NestedObject: schema.NestedAttributeObject{
												Attributes: map[string]schema.Attribute{
													"action": schema.ListAttribute{
														ElementType:         types.StringType,
														Optional:            true,
														Computed:            true,
														Description:         "the specific actions that will be allowed (Can be a string if only one element. A statement must have either Action or NotAction.)",
														MarkdownDescription: "the specific actions that will be allowed (Can be a string if only one element. A statement must have either Action or NotAction.)",
													},
													"effect": schema.StringAttribute{
														Optional:            true,
														Computed:            true,
														Description:         "the specific result of the statement (either an allow or an explicit deny)",
														MarkdownDescription: "the specific result of the statement (either an allow or an explicit deny)",
														Validators: []validator.String{
															stringvalidator.OneOf(
																"Allow",
																"Deny",
															),
														},
													},
													"not_action": schema.ListAttribute{
														ElementType:         types.StringType,
														Optional:            true,
														Computed:            true,
														Description:         "the specific exceptional actions (Can be a string if only one element. A statement must have either Action or NotAction.)",
														MarkdownDescription: "the specific exceptional actions (Can be a string if only one element. A statement must have either Action or NotAction.)",
													},
													"not_resource": schema.ListAttribute{
														ElementType:         types.StringType,
														Optional:            true,
														Computed:            true,
														Description:         "the objects that the statement does not cover (Can be a string if only one element. A statement must have either Resource or NotResource.)",
														MarkdownDescription: "the objects that the statement does not cover (Can be a string if only one element. A statement must have either Resource or NotResource.)",
													},
													"resource": schema.ListAttribute{
														ElementType:         types.StringType,
														Optional:            true,
														Computed:            true,
														Description:         "the objects that the statement covers (Can be a string if only one element. A statement must have either Resource or NotResource.)",
														MarkdownDescription: "the objects that the statement covers (Can be a string if only one element. A statement must have either Resource or NotResource.)",
													},
													"sid": schema.StringAttribute{
														Optional:            true,
														Computed:            true,
														Description:         "an optional identifier that you provide for the policy statement",
														MarkdownDescription: "an optional identifier that you provide for the policy statement",
													},
												},
											},
											Required:            true,
											Description:         "a list of group policy statements",
											MarkdownDescription: "a list of group policy statements",
										},
										"version": schema.StringAttribute{
											Optional:            true,
											Computed:            true,
											Description:         "S3 API Version (currently not used)",
											MarkdownDescription: "S3 API Version (currently not used)",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *groupsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *groupsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state groupsDataSourceModel
	var jsonData groupsDataSourceGolangModel
	var newDiags diag.Diagnostics

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &state)...)

	if resp.Diagnostics.HasError() {
		return
	}
	tflog.Debug(ctx, "1. Sending StorageGrid get request.")
	rp, _, _, err := d.client.SendRequest("GET", api_groups, nil, 200)

	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read example, got error: %s", err))
		return
	}

	// Map response body to model
	if err := json.Unmarshal(rp, &jsonData); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "2. Mapping response body to the TF state file.")
	for _, item := range jsonData.Data {
		var s3Sts []*S3PolicyStatementDataModel
		mgmtPolicies := &ManagementPolicyDataModel{
			ManageAllContainers:       types.BoolValue(item.Policies.Management.ManageAllContainers),
			ManageEndpoints:           types.BoolValue(item.Policies.Management.ManageEndpoints),
			ManageOwnContainerObjects: types.BoolValue(item.Policies.Management.ManageOwnContainerObjects),
			ManageOwnS3Credentials:    types.BoolValue(item.Policies.Management.ManageOwnS3Credentials),
			RootAccess:                types.BoolValue(item.Policies.Management.RootAccess),
		}
		for _, s3Pol := range item.Policies.S3.Statement {

			actions := []types.String{}
			notActions := []types.String{}
			resources := []types.String{}
			notResources := []types.String{}

			for _, action := range s3Pol.Action.AsStringSlice() {
				actions = append(actions, types.StringValue(action))
			}
			for _, notAction := range s3Pol.NotAction.AsStringSlice() {
				notActions = append(notActions, types.StringValue(notAction))
			}
			for _, resource := range s3Pol.Resource.AsStringSlice() {
				resources = append(resources, types.StringValue(resource))
			}
			for _, notResource := range s3Pol.NotResource.AsStringSlice() {
				notResources = append(notResources, types.StringValue(notResource))
			}
			s3Statement := &S3PolicyStatementDataModel{
				Sid:         types.StringValue(s3Pol.Sid),
				Effect:      types.StringValue(s3Pol.Effect),
				Action:      actions,
				NotAction:   notActions,
				Resource:    resources,
				NotResource: notResources,
			}
			s3Sts = append(s3Sts, s3Statement)
		}

		s3Policy := &S3PolicyDataModel{
			ID:        types.StringValue(item.Policies.S3.ID),
			Version:   types.StringValue(item.Policies.S3.Version),
			Statement: s3Sts,
		}

		groupData := &groupsDataSourceDataModel{
			ID:                 types.StringValue(item.ID),
			AccountID:          types.StringValue(item.AccountID),
			DisplayName:        types.StringValue(item.DisplayName),
			UniqueName:         types.StringValue(item.UniqueName),
			GroupURN:           types.StringValue(item.GroupURN),
			Federated:          types.BoolValue(item.Federated),
			ManagementReadOnly: types.BoolValue(item.ManagementReadOnly),
			Policies: &policiesDataModel{
				Management: mgmtPolicies,
				S3:         s3Policy,
			},
		}
		state.Data = append(state.Data, groupData)
	}

	resp.Diagnostics.Append(newDiags...)
	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "read all groups data source")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
