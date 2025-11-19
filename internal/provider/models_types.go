// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

/*
These are for creating HTTP client.
*/
type S3GridClient struct {
	address    string
	username   string
	password   string
	token      string
	tenant     string
	insecure   bool
	httpClient *http.Client
}

type S3GridClientJson struct {
	AccountId string `json:"accountId"`
	Username  string `json:"username"`
	Password  string `json:"password"`
	Cookie    bool   `json:"cookie"`
	CsrfToken bool   `json:"csrfToken"`
}

type S3GridClientReturnJson struct {
	Data string `json:"data"`
}

/*
These are for GET/POST Groups related data sources and resources
*/
type GroupsDataSourceDataModel struct {
	Data []*GroupsDataSourceModel `tfsdk:"data"`
}

type GroupsDataSourceModel struct {
	ID                 types.String   `tfsdk:"id"`
	AccountID          types.String   `tfsdk:"account_id"`
	DisplayName        types.String   `tfsdk:"display_name"`
	UniqueName         types.String   `tfsdk:"unique_name"`
	GroupURN           types.String   `tfsdk:"group_urn"`
	Federated          types.Bool     `tfsdk:"federated"`
	ManagementReadOnly types.Bool     `tfsdk:"management_read_only"`
	Policies           *PoliciesModel `tfsdk:"policies"`
}

type PoliciesModel struct {
	Management *ManagementPolicyDataModel `tfsdk:"management"`
	S3         *S3PolicyDataModel         `tfsdk:"s3"`
}

type ManagementPolicyDataModel struct {
	ManageAllContainers       types.Bool `tfsdk:"manage_all_containers"`
	ManageEndpoints           types.Bool `tfsdk:"manage_endpoints"`
	ManageOwnContainerObjects types.Bool `tfsdk:"manage_own_container_objects"`
	ManageOwnS3Credentials    types.Bool `tfsdk:"manage_own_s3_credentials"`
	ViewAllContainers         types.Bool `tfsdk:"view_all_containers"`
	RootAccess                types.Bool `tfsdk:"root_access"`
}

type S3PolicyDataModel struct {
	ID        types.String                  `tfsdk:"id"`
	Version   types.String                  `tfsdk:"version"`
	Statement []*S3PolicyStatementDataModel `tfsdk:"statement"`
}

type S3PolicyStatementDataModel struct {
	Sid         types.String   `tfsdk:"sid"`
	Effect      types.String   `tfsdk:"effect"`
	Action      []types.String `tfsdk:"action"`
	NotAction   []types.String `tfsdk:"not_action"`
	Resource    []types.String `tfsdk:"resource"`
	NotResource []types.String `tfsdk:"not_resource"`
}

type ResourceField struct {
	Resources []string
}

// Implement the UnmarshalJSON method to handle both string and []string
func (r *ResourceField) UnmarshalJSON(data []byte) error {
	// Try to unmarshal into a string first
	var singleResource string
	if err := json.Unmarshal(data, &singleResource); err == nil {
		r.Resources = []string{singleResource}
		return nil
	}

	// If it's not a string, try to unmarshal into a slice of strings
	var resourceList []string
	if err := json.Unmarshal(data, &resourceList); err == nil {
		r.Resources = resourceList
		return nil
	}

	// Return an error if neither case applies
	return fmt.Errorf("failed to unmarshal Resource field")
}

func (r ResourceField) AsStringSlice() []string {
	return r.Resources
}

type ActionField struct {
	Actions []string
}

func (a ActionField) AsStringSlice() []string {
	return a.Actions
}

// Implement the UnmarshalJSON method to handle both string and []string
func (r *ActionField) UnmarshalJSON(data []byte) error {
	// Try to unmarshal into a string first
	var singleResource string
	if err := json.Unmarshal(data, &singleResource); err == nil {
		r.Actions = []string{singleResource}
		return nil
	}

	// If it's not a string, try to unmarshal into a slice of strings
	var resourceList []string
	if err := json.Unmarshal(data, &resourceList); err == nil {
		r.Actions = resourceList
		return nil
	}

	// Return an error if neither case applies
	return fmt.Errorf("failed to unmarshal Resource field")
}

type Policies struct {
	Management ManagementPolicy `json:"management"`
	S3         S3Policy         `json:"s3"`
}

type ManagementPolicy struct {
	ManageAllContainers       bool `json:"manageAllContainers"`
	ManageEndpoints           bool `json:"manageEndpoints"`
	ManageOwnContainerObjects bool `json:"manageOwnContainerObjects"`
	ManageOwnS3Credentials    bool `json:"manageOwnS3Credentials"`
	ViewAllContainers         bool `json:"viewAllContainers"`
	RootAccess                bool `json:"rootAccess"`
}

type S3Policy struct {
	ID        string            `json:"id"`
	Version   string            `json:"version"`
	Statement []PolicyStatement `json:"Statement"`
}

type PolicyStatement struct {
	Sid         string        `json:"sid"`
	Effect      string        `json:"Effect"`
	Action      ActionField   `json:"Action"`
	NotAction   ActionField   `json:"NotAction"`
	Resource    ResourceField `json:"Resource"`    // Custom type to handle both string and []string for Resource
	NotResource ResourceField `json:"NotResource"` // Custom type to handle both string and []string for Resource
}

type GroupsDataObject struct {
	ID                 string   `json:"id"`
	AccountID          string   `json:"accountId"`
	DisplayName        string   `json:"displayName"`
	UniqueName         string   `json:"uniqueName"`
	GroupURN           string   `json:"groupURN"`
	Federated          bool     `json:"federated"`
	ManagementReadOnly bool     `json:"managementReadOnly"`
	Policies           Policies `json:"policies"`
}

type groupsDataSourceGolangModel struct {
	Data []GroupsDataObject `json:"data"`
}

type groupsDataSourceGolangModelSingle struct {
	Data GroupsDataObject `json:"data"`
}

/*
These are for creating new group (as part of resources)
*/
type GroupPostPolicies struct {
	Management ManagementPolicy `json:"management"`
	S3         S3PostPolicy     `json:"s3"`
}

type S3PostPolicy struct {
	ID        string                     `json:"Id"`
	Version   string                     `json:"Version"`
	Statement []GroupPostPolicyStatement `json:"Statement"`
}

type GroupPostPolicyStatement struct {
	Sid         string   `json:"Sid"`
	Effect      string   `json:"Effect"`
	Action      []string `json:"Action,omitempty"`
	NotAction   []string `json:"NotAction,omitempty"`
	Resource    []string `json:"Resource,omitempty"`
	NotResource []string `json:"NotResource,omitempty"`
}

type GroupsPostDataObject struct {
	DisplayName        string            `json:"displayName"`
	UniqueName         string            `json:"uniqueName"`
	ManagementReadOnly bool              `json:"managementReadOnly"`
	Policies           GroupPostPolicies `json:"policies"`
}

/*
These are for creating or fetching users data
*/
type UserModel struct {
	UniqueName string   `json:"uniqueName"`
	FullName   string   `json:"fullName"`
	MemberOf   []string `json:"memberOf"`
	Disable    bool     `json:"disable"`
	AccountId  string   `json:"accountId"`
	ID         string   `json:"id"`
	Federated  bool     `json:"federated"`
	UserURN    string   `json:"userURN"`
}

type UserModelPostRequest struct {
	UniqueName string   `json:"uniqueName"`
	FullName   string   `json:"fullName"`
	MemberOf   []string `json:"memberOf"`
	Disable    bool     `json:"disable"`
}

type UsersDataModelSingle struct {
	Data UserModel `json:"data"`
}

type UsersDataModel struct {
	Data []UserModel `json:"data"`
}

type usersDataSourceModel struct {
	Data []*usersDataSourceDataModel `tfsdk:"data"`
}

type usersDataSourceDataModel struct {
	UniqueName types.String   `tfsdk:"unique_name"`
	FullName   types.String   `tfsdk:"full_name"`
	MemberOf   []types.String `tfsdk:"member_of"`
	Disable    types.Bool     `tfsdk:"disable"`
	AccountId  types.String   `tfsdk:"account_id"`
	ID         types.String   `tfsdk:"id"`
	Federated  types.Bool     `tfsdk:"federated"`
	UserURN    types.String   `tfsdk:"user_urn"`
}

/*
These are for creating or fetching S3 access/secret keys
*/
type UserIDS3AccessKeys struct {
	Data []S3AccessKey `json:"data"`
}

type UserIDS3AccessKeySingle struct {
	Data S3AccessKey `json:"data"`
}

type S3AccessKey struct {
	ID          string `json:"id"`
	AccountId   string `json:"accountId"`
	DisplayName string `json:"displayName"`
	UserURN     string `json:"userURN"`
	UserUUID    string `json:"userUUID"`
	Expires     string `json:"expires"`
}

type UserIDS3AccessSecretKeySingle struct {
	Data S3AccessSecretKey `json:"data"`
}

type S3AccessSecretKey struct {
	ID              string `json:"id"`
	AccountId       string `json:"accountId"`
	DisplayName     string `json:"displayName"`
	UserURN         string `json:"userURN"`
	UserUUID        string `json:"userUUID"`
	Expires         string `json:"expires"`
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
}

type UserIDS3AccessSecretKeysCreateJson struct {
	Expires *string `json:"expires"`
}

type UserIDS3AllKeysModel struct {
	UserUUID types.String        `tfsdk:"user_uuid"`
	Data     []*S3AccessKeyModel `tfsdk:"data"`
}

type UserIDS3AccessKeysModel struct {
	UserUUID  types.String      `tfsdk:"user_uuid"`
	AccessKey types.String      `tfsdk:"access_key"`
	Data      *S3AccessKeyModel `tfsdk:"data"`
}

type S3AccessKeyModel struct {
	ID          types.String `tfsdk:"id"`
	AccountId   types.String `tfsdk:"account_id"`
	DisplayName types.String `tfsdk:"display_name"`
	UserURN     types.String `tfsdk:"user_urn"`
	UserUUID    types.String `tfsdk:"user_uuid"`
	Expires     types.String `tfsdk:"expires"`
}

type UserIDS3AccessSecretKeysModel struct {
	UserUUID types.String              `tfsdk:"user_uuid"`
	Expires  types.String              `tfsdk:"expires"`
	Data     *S3AccessKeyResourceModel `tfsdk:"data"`
}

type S3AccessKeyResourceModel struct {
	ID              types.String `tfsdk:"id"`
	AccountId       types.String `tfsdk:"account_id"`
	DisplayName     types.String `tfsdk:"display_name"`
	UserURN         types.String `tfsdk:"user_urn"`
	UserUUID        types.String `tfsdk:"user_uuid"`
	Expires         types.String `tfsdk:"expires"`
	AccessKey       types.String `tfsdk:"access_key"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
}

type BucketResourceModel struct {
	Name   types.String `tfsdk:"name"`
	Region types.String `tfsdk:"region"`
}

func (m *BucketResourceModel) ToBucketModel() BucketApiRequestModel {
	return BucketApiRequestModel{
		Name:   m.Name.ValueString(),
		Region: m.Region.ValueString(),
	}
}

type BucketApiRequestModel struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}
type BucketApiResponseModel struct {
	Data BucketApiRequestModel `json:"data"`
}
