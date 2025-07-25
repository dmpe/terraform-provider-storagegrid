// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var (
	_ resource.Resource                = &groupsResource{}
	_ resource.ResourceWithConfigure   = &groupsResource{}
	_ resource.ResourceWithImportState = &groupsResource{}
)

func NewGroupsResource() resource.Resource {
	return &groupsResource{}
}

// NewGroupsResource defines the resource implementation.
type groupsResource struct {
	client *S3GridClient
}

func (r *groupsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_groups"
}

func (r *groupsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	defaultEmptyTagList, _ := basetypes.NewListValue(types.StringType, []attr.Value{})

	resp.Schema = schema.Schema{
		MarkdownDescription: "Create new group - a resource",
		Attributes: map[string]schema.Attribute{
			"group_urn": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Contains the Group uniqueName and Account ID (generated automatically)",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"federated": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.UseStateForUnknown(),
				},
				Description: "True if the Group is federated, for example, an LDAP Group",
			},
			id: schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"account_id": schema.StringAttribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"display_name": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					// Must contain at least 1 and no more than 32 characters
					stringvalidator.LengthAtLeast(1),
					stringvalidator.LengthAtMost(32),
				},
			},
			"unique_name": schema.StringAttribute{
				Required: true,
				Description: `The unique name for a group, which cannot be changed. 
For federated group, the unique name comes from the identity source. The value to specify depends on the type of identity source in use:
Active Directory: Use the sAMAccountName attribute.
OpenLDAP: Use the CN (Common Name).
Other LDAP: Identify the appropriate unique name value for the LDAP server.`,
			},
			"management_read_only": schema.BoolAttribute{
				Description: "Select whether users can change settings and perform operations or whether they can only view settings and features.",
				Required:    true,
			},
			"policies": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"management": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"manage_all_containers": schema.BoolAttribute{
								Optional:    true,
								Description: "Allows users to change settings of all S3 buckets (or Swift containers) in this account. Supersedes the View all buckets permission. Applies to the Tenant Manager UI and API only and does not affect the permissions granted by an S3 group policy.",
							},
							"manage_endpoints": schema.BoolAttribute{
								Optional:    true,
								Description: "Allows users to configure endpoints for platform services.",
							},
							"manage_own_container_objects": schema.BoolAttribute{
								Optional:    true,
								Description: "When combined with the View all buckets or Manage all buckets permission, allows users to view and manage objects from the S3 Console tab on the details page for a bucket.",
							},
							"manage_own_s3_credentials": schema.BoolAttribute{
								Optional:    true,
								Description: "Allows users to create and delete their own S3 access keys.",
							},
							"view_all_containers": schema.BoolAttribute{
								Optional:    true,
								Description: "Allows users to view settings of all S3 buckets (or Swift containers) in this account. Superseded by the Manage all buckets permission. Applies to the Tenant Manager UI and API only and does not affect the permissions granted by an S3 group policy.",
							},
							"root_access": schema.BoolAttribute{
								Description: "Allows users to access all administration features. Root access permission supersedes all other permissions.",
								Optional:    true,
							},
						},
					},
					"s3": schema.SingleNestedAttribute{
						Required: true,
						Attributes: map[string]schema.Attribute{
							"statement": schema.ListNestedAttribute{
								NestedObject: schema.NestedAttributeObject{
									Attributes: map[string]schema.Attribute{
										act: schema.ListAttribute{
											ElementType:         types.StringType,
											Optional:            true,
											Description:         "the specific actions that will be allowed (Can be a string if only one element. A statement must have either Action or NotAction.)",
											MarkdownDescription: "the specific actions that will be allowed (Can be a string if only one element. A statement must have either Action or NotAction.)",
											Validators: []validator.List{
												listvalidator.AtLeastOneOf(
													path.MatchRelative().AtParent().AtName(act),
													path.MatchRelative().AtParent().AtName(n_act),
												),
											},
										},
										"effect": schema.StringAttribute{
											Optional: true,
											// Computed:            true,
											Description:         "the specific result of the statement (either an allow or an explicit deny)",
											MarkdownDescription: "the specific result of the statement (either an allow or an explicit deny)",
											Validators: []validator.String{
												stringvalidator.OneOf(
													"Allow",
													"Deny",
												),
											},
										},
										n_act: schema.ListAttribute{
											ElementType:         types.StringType,
											Optional:            true,
											Computed:            true,
											Description:         "the specific exceptional actions (Can be a string if only one element. A statement must have either Action or NotAction.)",
											MarkdownDescription: "the specific exceptional actions (Can be a string if only one element. A statement must have either Action or NotAction.)",
											Validators: []validator.List{
												listvalidator.AtLeastOneOf(
													path.MatchRelative().AtParent().AtName(act),
													path.MatchRelative().AtParent().AtName(n_act),
												),
											},
											Default: listdefault.StaticValue(defaultEmptyTagList),
										},
										n_res: schema.ListAttribute{
											ElementType:         types.StringType,
											Optional:            true,
											Computed:            true,
											Description:         "the objects that the statement does not cover (Can be a string if only one element. A statement must have either Resource or NotResource.)",
											MarkdownDescription: "the objects that the statement does not cover (Can be a string if only one element. A statement must have either Resource or NotResource.)",
											Validators: []validator.List{
												listvalidator.AtLeastOneOf(
													path.MatchRelative().AtParent().AtName(n_res),
													path.MatchRelative().AtParent().AtName(res),
												),
											},
											Default: listdefault.StaticValue(defaultEmptyTagList),
										},
										res: schema.ListAttribute{
											ElementType:         types.StringType,
											Optional:            true,
											Description:         "the objects that the statement covers (Can be a string if only one element. A statement must have either Resource or NotResource.)",
											MarkdownDescription: "the objects that the statement covers (Can be a string if only one element. A statement must have either Resource or NotResource.)",
											Validators: []validator.List{
												listvalidator.AtLeastOneOf(
													path.MatchRelative().AtParent().AtName(n_res),
													path.MatchRelative().AtParent().AtName(res),
												),
											},
										},
										"sid": schema.StringAttribute{
											Optional:            true,
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
								Computed:            true,
								Optional:            true,
								Description:         "S3 API Version (currently not used)",
								MarkdownDescription: "S3 API Version (currently not used)",
								Default:             stringdefault.StaticString("2006-03-01"),
							},
							id: schema.StringAttribute{
								Computed:            true,
								Optional:            true,
								Description:         "S3 Policy ID provided by policy generator tools (currently not used)",
								MarkdownDescription: "S3 Policy ID provided by policy generator tools (currently not used)",
								Default:             stringdefault.StaticString("terraform-storagegrid-s3"),
							},
						},
					},
				},
			},
		},
	}
}

func (r *groupsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*S3GridClient)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *groupsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GroupsDataSourceModel
	var s3Sts []GroupPostPolicyStatement
	var returnBody groupsDataSourceGolangModelSingle
	var returnS3Sts []*S3PolicyStatementDataModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "1. Create to json body and fill it with the passed variables.")
	mgmtPolicies := &ManagementPolicy{
		ManageAllContainers:       plan.Policies.Management.ManageAllContainers.ValueBool(),
		ManageEndpoints:           plan.Policies.Management.ManageEndpoints.ValueBool(),
		ManageOwnContainerObjects: plan.Policies.Management.ManageOwnContainerObjects.ValueBool(),
		ManageOwnS3Credentials:    plan.Policies.Management.ManageOwnS3Credentials.ValueBool(),
		ViewAllContainers:         plan.Policies.Management.ViewAllContainers.ValueBool(),
		RootAccess:                plan.Policies.Management.RootAccess.ValueBool(),
	}

	for _, s3Pol := range plan.Policies.S3.Statement {
		actions := []string{}
		notActions := []string{}
		resources := []string{}
		notResources := []string{}

		for _, action := range s3Pol.Action {
			actions = append(actions, action.ValueString())
		}
		for _, notAction := range s3Pol.NotAction {
			notActions = append(notActions, notAction.ValueString())
		}
		for _, resource := range s3Pol.Resource {
			resources = append(resources, resource.ValueString())
		}
		for _, notResource := range s3Pol.NotResource {
			notResources = append(notResources, notResource.ValueString())
		}
		s3Statement := &GroupPostPolicyStatement{
			Sid:         s3Pol.Sid.ValueString(),
			Effect:      s3Pol.Effect.ValueString(),
			Action:      actions,
			NotAction:   notActions,
			Resource:    resources,
			NotResource: notResources,
		}
		s3Sts = append(s3Sts, *s3Statement)
	}

	s3Policy := &S3PostPolicy{
		ID:        plan.Policies.S3.ID.ValueString(),
		Version:   plan.Policies.S3.Version.ValueString(),
		Statement: s3Sts,
	}

	body := &GroupsPostDataObject{
		DisplayName:        plan.DisplayName.ValueString(),
		UniqueName:         plan.UniqueName.ValueString(),
		ManagementReadOnly: plan.ManagementReadOnly.ValueBool(),
		Policies: GroupPostPolicies{
			Management: *mgmtPolicies,
			S3:         *s3Policy,
		},
	}

	tflog.Debug(ctx, "2. Execute Request against REST api.")
	httpResp, _, _, err := r.client.SendRequest("POST", api_groups, body, 201)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create example, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Group has been created and now we unmarshal it to json object.")
	if err := json.Unmarshal(httpResp, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "4. Mapping json body back to the state file.")
	plan.ID = types.StringValue(returnBody.Data.ID)
	plan.AccountID = types.StringValue(returnBody.Data.AccountID)
	plan.DisplayName = types.StringValue(returnBody.Data.DisplayName)
	plan.UniqueName = types.StringValue(returnBody.Data.UniqueName)
	plan.GroupURN = types.StringValue(returnBody.Data.GroupURN)
	plan.Federated = types.BoolValue(returnBody.Data.Federated)
	plan.ManagementReadOnly = types.BoolValue(returnBody.Data.ManagementReadOnly)

	returnMgmtPolicies := &ManagementPolicyDataModel{
		ManageAllContainers:       types.BoolValue(returnBody.Data.Policies.Management.ManageAllContainers),
		ManageEndpoints:           types.BoolValue(returnBody.Data.Policies.Management.ManageEndpoints),
		ManageOwnContainerObjects: types.BoolValue(returnBody.Data.Policies.Management.ManageOwnContainerObjects),
		ManageOwnS3Credentials:    types.BoolValue(returnBody.Data.Policies.Management.ManageOwnS3Credentials),
		ViewAllContainers:         types.BoolValue(returnBody.Data.Policies.Management.ViewAllContainers),
		RootAccess:                types.BoolValue(returnBody.Data.Policies.Management.RootAccess),
	}
	for _, returnS3Pol := range returnBody.Data.Policies.S3.Statement {
		actions := []types.String{}
		notActions := []types.String{}
		resources := []types.String{}
		notResources := []types.String{}

		for _, action := range returnS3Pol.Action.AsStringSlice() {
			actions = append(actions, types.StringValue(action))
		}
		for _, notAction := range returnS3Pol.NotAction.AsStringSlice() {
			notActions = append(notActions, types.StringValue(notAction))
		}
		for _, resource := range returnS3Pol.Resource.AsStringSlice() {
			resources = append(resources, types.StringValue(resource))
		}
		for _, notResource := range returnS3Pol.NotResource.AsStringSlice() {
			notResources = append(notResources, types.StringValue(notResource))
		}
		returnS3Statement := &S3PolicyStatementDataModel{
			Sid:         types.StringValue(returnS3Pol.Sid),
			Effect:      types.StringValue(returnS3Pol.Effect),
			Action:      actions,
			NotAction:   notActions,
			Resource:    resources,
			NotResource: notResources,
		}
		returnS3Sts = append(returnS3Sts, returnS3Statement)
	}

	returnS3Policy := &S3PolicyDataModel{
		ID:        types.StringValue(returnBody.Data.Policies.S3.ID),
		Version:   types.StringValue(returnBody.Data.Policies.S3.Version),
		Statement: returnS3Sts,
	}
	newPolicies := &PoliciesModel{
		Management: returnMgmtPolicies,
		S3:         returnS3Policy,
	}

	plan.Policies = newPolicies

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a new group")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *groupsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state GroupsDataSourceModel
	var returnBody groupsDataSourceGolangModelSingle
	var returnS3Sts []*S3PolicyStatementDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "1. Get refreshed group information.")
	respBody, _, respCode, err := r.client.SendRequest("GET", api_groups+"/"+state.ID.ValueString(), nil, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading StorageGrid Group",
			"Could not read StorageGrid group ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "2. Unmarshal group information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Overwrite fields with refreshed information.")
	state.ID = types.StringValue(returnBody.Data.ID)
	state.AccountID = types.StringValue(returnBody.Data.AccountID)
	state.DisplayName = types.StringValue(returnBody.Data.DisplayName)
	state.UniqueName = types.StringValue(returnBody.Data.UniqueName)
	state.GroupURN = types.StringValue(returnBody.Data.GroupURN)
	state.Federated = types.BoolValue(returnBody.Data.Federated)
	state.ManagementReadOnly = types.BoolValue(returnBody.Data.ManagementReadOnly)

	returnMgmtPolicies := &ManagementPolicyDataModel{
		ManageAllContainers:       types.BoolValue(returnBody.Data.Policies.Management.ManageAllContainers),
		ManageEndpoints:           types.BoolValue(returnBody.Data.Policies.Management.ManageEndpoints),
		ManageOwnContainerObjects: types.BoolValue(returnBody.Data.Policies.Management.ManageOwnContainerObjects),
		ManageOwnS3Credentials:    types.BoolValue(returnBody.Data.Policies.Management.ManageOwnS3Credentials),
		ViewAllContainers:         types.BoolValue(returnBody.Data.Policies.Management.ViewAllContainers),
		RootAccess:                types.BoolValue(returnBody.Data.Policies.Management.RootAccess),
	}
	for _, returnS3Pol := range returnBody.Data.Policies.S3.Statement {
		actions := []types.String{}
		notActions := []types.String{}
		resources := []types.String{}
		notResources := []types.String{}

		for _, action := range returnS3Pol.Action.AsStringSlice() {
			actions = append(actions, types.StringValue(action))
		}
		for _, notAction := range returnS3Pol.NotAction.AsStringSlice() {
			notActions = append(notActions, types.StringValue(notAction))
		}
		for _, resource := range returnS3Pol.Resource.AsStringSlice() {
			resources = append(resources, types.StringValue(resource))
		}
		for _, notResource := range returnS3Pol.NotResource.AsStringSlice() {
			notResources = append(notResources, types.StringValue(notResource))
		}
		returnS3Statement := &S3PolicyStatementDataModel{
			Sid:         types.StringValue(returnS3Pol.Sid),
			Effect:      types.StringValue(returnS3Pol.Effect),
			Action:      actions,
			NotAction:   notActions,
			Resource:    resources,
			NotResource: notResources,
		}
		returnS3Sts = append(returnS3Sts, returnS3Statement)
	}

	returnS3Policy := &S3PolicyDataModel{
		ID:        types.StringValue(returnBody.Data.Policies.S3.ID),
		Version:   types.StringValue(returnBody.Data.Policies.S3.Version),
		Statement: returnS3Sts,
	}
	newPolicies := &PoliciesModel{
		Management: returnMgmtPolicies,
		S3:         returnS3Policy,
	}

	state.Policies = newPolicies

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *groupsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state GroupsDataSourceModel
	var plan GroupsDataSourceModel
	var returnBody groupsDataSourceGolangModelSingle
	var s3Sts []GroupPostPolicyStatement
	var returnS3Sts []*S3PolicyStatementDataModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var groupID = state.ID.ValueString()

	tflog.Debug(ctx, "1. Create updated group information.")
	mgmtPolicies := &ManagementPolicy{
		ManageAllContainers:       plan.Policies.Management.ManageAllContainers.ValueBool(),
		ManageEndpoints:           plan.Policies.Management.ManageEndpoints.ValueBool(),
		ManageOwnContainerObjects: plan.Policies.Management.ManageOwnContainerObjects.ValueBool(),
		ManageOwnS3Credentials:    plan.Policies.Management.ManageOwnS3Credentials.ValueBool(),
		ViewAllContainers:         plan.Policies.Management.ViewAllContainers.ValueBool(),
		RootAccess:                plan.Policies.Management.RootAccess.ValueBool(),
	}

	for _, s3Pol := range plan.Policies.S3.Statement {
		actions := []string{}
		notActions := []string{}
		resources := []string{}
		notResources := []string{}

		for _, action := range s3Pol.Action {
			actions = append(actions, action.ValueString())
		}
		for _, notAction := range s3Pol.NotAction {
			notActions = append(notActions, notAction.ValueString())
		}
		for _, resource := range s3Pol.Resource {
			resources = append(resources, resource.ValueString())
		}
		for _, notResource := range s3Pol.NotResource {
			notResources = append(notResources, notResource.ValueString())
		}
		s3Statement := &GroupPostPolicyStatement{
			Sid:         s3Pol.Sid.ValueString(),
			Effect:      s3Pol.Effect.ValueString(),
			Action:      actions,
			NotAction:   notActions,
			Resource:    resources,
			NotResource: notResources,
		}
		s3Sts = append(s3Sts, *s3Statement)
	}

	s3Policy := &S3PostPolicy{
		ID:        plan.Policies.S3.ID.ValueString(),
		Version:   plan.Policies.S3.Version.ValueString(),
		Statement: s3Sts,
	}

	body := &GroupsPostDataObject{
		DisplayName:        plan.DisplayName.ValueString(),
		UniqueName:         plan.UniqueName.ValueString(),
		ManagementReadOnly: plan.ManagementReadOnly.ValueBool(),
		Policies: GroupPostPolicies{
			Management: *mgmtPolicies,
			S3:         *s3Policy,
		},
	}

	tflog.Debug(ctx, "2. Execute Request against REST api.")
	_, _, _, err := r.client.SendRequest("PUT", api_groups+"/"+groupID, body, 200)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update group information, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "3. Get refreshed group information.")
	respBody, _, respCode, err := r.client.SendRequest("GET", api_groups+"/"+groupID, nil, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error Reading StorageGrid Group",
			"Could not read StorageGrid group ID "+groupID+": "+err.Error(),
		)
		return
	}

	tflog.Debug(ctx, "4. Unmarshal group information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to parse response, got error: %s", err))
		return
	}

	tflog.Debug(ctx, "5. Overwrite fields with refreshed information.")
	state.ID = types.StringValue(returnBody.Data.ID)
	state.AccountID = types.StringValue(returnBody.Data.AccountID)
	state.DisplayName = types.StringValue(returnBody.Data.DisplayName)
	state.UniqueName = types.StringValue(returnBody.Data.UniqueName)
	state.GroupURN = types.StringValue(returnBody.Data.GroupURN)
	state.Federated = types.BoolValue(returnBody.Data.Federated)
	state.ManagementReadOnly = types.BoolValue(returnBody.Data.ManagementReadOnly)

	returnMgmtPolicies := &ManagementPolicyDataModel{
		ManageAllContainers:       types.BoolValue(returnBody.Data.Policies.Management.ManageAllContainers),
		ManageEndpoints:           types.BoolValue(returnBody.Data.Policies.Management.ManageEndpoints),
		ManageOwnContainerObjects: types.BoolValue(returnBody.Data.Policies.Management.ManageOwnContainerObjects),
		ManageOwnS3Credentials:    types.BoolValue(returnBody.Data.Policies.Management.ManageOwnS3Credentials),
		ViewAllContainers:         types.BoolValue(returnBody.Data.Policies.Management.ViewAllContainers),
		RootAccess:                types.BoolValue(returnBody.Data.Policies.Management.RootAccess),
	}
	for _, returnS3Pol := range returnBody.Data.Policies.S3.Statement {
		actions := []types.String{}
		notActions := []types.String{}
		resources := []types.String{}
		notResources := []types.String{}

		for _, action := range returnS3Pol.Action.AsStringSlice() {
			actions = append(actions, types.StringValue(action))
		}
		for _, notAction := range returnS3Pol.NotAction.AsStringSlice() {
			notActions = append(notActions, types.StringValue(notAction))
		}
		for _, resource := range returnS3Pol.Resource.AsStringSlice() {
			resources = append(resources, types.StringValue(resource))
		}
		for _, notResource := range returnS3Pol.NotResource.AsStringSlice() {
			notResources = append(notResources, types.StringValue(notResource))
		}
		returnS3Statement := &S3PolicyStatementDataModel{
			Sid:         types.StringValue(returnS3Pol.Sid),
			Effect:      types.StringValue(returnS3Pol.Effect),
			Action:      actions,
			NotAction:   notActions,
			Resource:    resources,
			NotResource: notResources,
		}
		returnS3Sts = append(returnS3Sts, returnS3Statement)
	}

	returnS3Policy := &S3PolicyDataModel{
		ID:        types.StringValue(returnBody.Data.Policies.S3.ID),
		Version:   types.StringValue(returnBody.Data.Policies.S3.Version),
		Statement: returnS3Sts,
	}
	newPolicies := &PoliciesModel{
		Management: returnMgmtPolicies,
		S3:         returnS3Policy,
	}

	state.Policies = newPolicies

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *groupsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GroupsDataSourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// in order for us to delete it, we first need to retrieve the same group and its ID
	_, _, _, err := r.client.SendRequest("DELETE", api_groups+"/"+state.ID.ValueString(), nil, 204)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting StorageGrid group",
			"Could not delete order, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *groupsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
