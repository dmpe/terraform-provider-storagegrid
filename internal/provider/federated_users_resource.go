// Copyright github.com/dmpe 2024, 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider-defined types fully satisfy framework interfaces.
var (
	_ resource.Resource              = &federatedUsersResource{}
	_ resource.ResourceWithConfigure = &federatedUsersResource{}
)

// NewFederatedUsersResource returns a new resource instance.
func NewFederatedUsersResource() resource.Resource {
	return &federatedUsersResource{}
}

// federatedUsersResource defines the resource implementation.
type federatedUsersResource struct {
	client *S3GridClient
}

func (r *federatedUsersResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_federated_users"
}

func (r *federatedUsersResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Import a federated user into the StorageGrid.

Use this resource to automate the import of a federated user into the StorageGrid to create dependent resources such as
access keys etc.

**Workflow**
- If the federated user does not yet exist in StorageGRID, the user will be imported to the StorageGRID when creating this resource.
- If the user does already exist in the StorageGRID, the available user data will be read to this resource during creation.

> [!warning] Only available in StorageGRID 12 or higher
> This resource is only available when working with StorageGRID 12 or higher.
> Targeting earlier StorageGRID versions will result in errors during the Terraform apply step.
`,
		Attributes: map[string]schema.Attribute{
			unique_name: schema.StringAttribute{
				Required: true,
				MarkdownDescription: `The unique name of the user to import - must match the name of the user in the identity provider.

The unique name must be prefixed with 'federated-user', such that the resulting unique name is of the form 'federated-user/<name>'.
`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			fl_name: schema.StringAttribute{
				Description: "The human-readable name for the User",
				Computed:    true,
			},
			"member_of": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Group memberships for this User",
			},
			"user_urn": schema.StringAttribute{
				Computed: true,
			},
			"account_id": schema.StringAttribute{
				Computed: true,
			},
			id: schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (r *federatedUsersResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	usesSupportedVersion, err := r.verifyProductVersion()
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Resource Configure Error", err.Error())
		return
	}

	if !usesSupportedVersion {
		resp.Diagnostics.AddError("Unsupported Product Version", "Please upgrade to StorageGrid 12 or newer.")
		return
	}
}

func (r *federatedUsersResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan federatedUsersResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created := r.refreshUser(plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || created == nil {
		return
	}

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "imported a federated user")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, created)...)
}

func (r *federatedUsersResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state federatedUsersResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	read := r.refreshUser(state, &resp.Diagnostics)
	if resp.Diagnostics.HasError() || read == nil {
		return
	}

	stateMemberOf := state.memberOfAsStringList(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}
	readMemberOf := read.memberOfAsStringList(ctx, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// if `member_of` attribute value is semantically equal to the current state, do not update it
	if EqualElements(readMemberOf, stateMemberOf) {
		read.MemberOf = state.MemberOf
	}

	// Set the refreshed state
	resp.Diagnostics.Append(resp.State.Set(ctx, read)...)
}

func (r *federatedUsersResource) Update(_ context.Context, _ resource.UpdateRequest, _ *resource.UpdateResponse) {
	// no-op
	// Changing the unique name forces a re-creation of the resource as a different federated user will then be
	// imported into the StorageGrid.
}

func (r *federatedUsersResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state federatedUsersResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	userID := state.ID.ValueString()
	if userID == "" {
		returnBody, respCode, err := r.readUser("", state.UniqueName.ValueString())
		if err != nil {
			if respCode == http.StatusNotFound {
				return
			}
			resp.Diagnostics.AddError(
				"Error Resolving StorageGrid user",
				"Could not resolve user ID for deletion, unexpected error: "+err.Error(),
			)
			return
		}
		if returnBody.Data.ID == "" {
			resp.Diagnostics.AddError("Error Resolving StorageGrid user", "Could not resolve user ID for deletion")
			return
		}
		userID = returnBody.Data.ID
	}

	// in order for us to delete it, we first need to retrieve the same user and its ID
	_, _, _, err := r.client.SendRequest("DELETE", api_users+"/"+userID, nil, 204)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting StorageGrid user",
			"Could not delete user, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *federatedUsersResource) refreshUser(plan federatedUsersResourceModel, diags *diag.Diagnostics) *federatedUsersResourceModel {
	body := struct {
		UsernameOrUuid string `json:"usernameOrUUID"`
	}{
		UsernameOrUuid: strings.ReplaceAll(plan.UniqueName.ValueString(), "federated-user/", ""),
	}

	_, _, _, err := r.client.SendRequest("POST", "/org/import-federated-user", body, 204)
	if err != nil {
		diags.AddError("Client Error", fmt.Sprintf("Unable to import federated user '%s', got error: %s", plan.UniqueName.ValueString(), err))
		return nil
	}

	read, responseCode, err := r.readUser("", plan.UniqueName.ValueString())
	if err != nil {
		if responseCode == http.StatusNotFound {
			diags.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
			return nil
		}

		diags.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return nil
	}

	plan.update(read)

	return &plan
}

func (r *federatedUsersResource) readUser(userID string, uniqueName string) (UsersDataModelSingle, int, error) {
	var returnBody UsersDataModelSingle

	fullPath := api_users + "/" + userID
	if userID == "" {
		if uniqueName == "" {
			return returnBody, 0, fmt.Errorf("cannot read StorageGrid user without id or unique_name")
		}
		fullPath = api_users + "/" + uniqueName
	}

	respBody, _, respCode, err := r.client.SendRequest("GET", fullPath, nil, 200)
	if err != nil {
		return returnBody, respCode, err
	}

	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		return returnBody, respCode, err
	}

	return returnBody, respCode, nil
}

func (r *federatedUsersResource) verifyProductVersion() (bool, error) {
	httpResponse, _, _, err := r.client.SendRequest("GET", "/org/config/product-version", nil, 200)
	if err != nil {
		return false, err
	}

	var productVersion struct {
		Data struct {
			ProductVersion string `json:"productVersion"`
		} `json:"data"`
	}

	if err := json.Unmarshal(httpResponse, &productVersion); err != nil {
		return false, err
	}

	versionParts := strings.Split(productVersion.Data.ProductVersion, ".")
	if len(versionParts) == 0 {
		return false, fmt.Errorf("invalid product version format")
	}
	majorVersion, err := strconv.Atoi(versionParts[0])
	if err != nil {
		return false, fmt.Errorf("invalid major version: %w", err)
	}
	return majorVersion >= 12, nil
}

type federatedUsersResourceModel struct {
	UniqueName types.String `tfsdk:"unique_name"`
	FullName   types.String `tfsdk:"full_name"`
	MemberOf   types.List   `tfsdk:"member_of"`
	AccountId  types.String `tfsdk:"account_id"`
	ID         types.String `tfsdk:"id"`
	UserURN    types.String `tfsdk:"user_urn"`
}

func (f *federatedUsersResourceModel) update(in UsersDataModelSingle) {
	f.ID = types.StringValue(in.Data.ID)
	f.UniqueName = types.StringValue(in.Data.UniqueName)
	f.FullName = types.StringValue(in.Data.FullName)
	f.AccountId = types.StringValue(in.Data.AccountId)
	f.UserURN = types.StringValue(in.Data.UserURN)
	f.AccountId = types.StringValue(in.Data.AccountId)

	members := make([]attr.Value, len(in.Data.MemberOf))
	for i, member := range in.Data.MemberOf {
		members[i] = types.StringValue(member)
	}
	f.MemberOf = types.ListValueMust(types.StringType, members)
}

func (f *federatedUsersResourceModel) memberOfAsStringList(ctx context.Context, diags *diag.Diagnostics) []string {
	elements := make([]types.String, 0, len(f.MemberOf.Elements()))
	diags.Append(f.MemberOf.ElementsAs(ctx, &elements, false)...)
	if diags.HasError() {
		return nil
	}

	var members []string
	for _, member := range elements {
		members = append(members, member.ValueString())
	}
	return members
}
