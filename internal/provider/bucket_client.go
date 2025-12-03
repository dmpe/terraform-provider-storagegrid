// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// NewBucketClient returns a new instance of BucketClient configured with the given S3GridClient as the underlying API
// client.
func NewBucketClient(apiClient *S3GridClient) *BucketClient {
	return &BucketClient{apiClient: apiClient}
}

// BucketClient is a dedicated API client for interacting with StorageGrid buckets.
// It implements the CRUD operations for buckets required by Terraform Resources and Data Sources.
type BucketClient struct {
	apiClient *S3GridClient
}

// Create creates a new StorageGrid bucket from the given BucketResourceModel configuration.
func (c *BucketClient) Create(ctx context.Context, bucket BucketResourceModel) (*BucketResourceModel, error) {
	httpResp, _, _, err := c.apiClient.SendRequest("POST", api_buckets, bucket.ToBucketModel(), 201)
	if err != nil {
		return nil, fmt.Errorf("unable to create StorageGrid container: %w", err)
	}

	var returnBody BucketApiResponseModel
	if err := json.Unmarshal(httpResp, &returnBody); err != nil {
		return nil, fmt.Errorf("unable to unmarshal create StorageGrid container response: %w", err)
	}

	// if no region is provided, we need to read the resource to get the used default region.
	read, err := c.Read(ctx, returnBody.Data.Name)
	if err != nil {
		return nil, fmt.Errorf("unable to read newly created StorageGrid container: %w", err)
	}

	return read, nil
}

// Update updates the StorageGrid bucket with the given BucketResourceModel configuration.
// It takes the current state of the configuration as well as the planned changes as input to ensure no changes are made
// to the object lock configuration, because this is not allowed by the API.
func (c *BucketClient) Update(ctx context.Context, plan, state BucketResourceModel) (*BucketResourceModel, error) {
	if (plan.ObjectLockConfiguration != nil && state.ObjectLockConfiguration == nil) || (plan.ObjectLockConfiguration == nil && state.ObjectLockConfiguration != nil) {
		return nil, fmt.Errorf("object Lock Configuration cannot be changed once set")
	}

	payload := plan.ToBucketModel()

	if payload.S3ObjectLock == nil {
		return &state, nil
	}

	_, _, _, err := c.apiClient.SendRequest("PUT", fmt.Sprintf("%s/%s/object-lock", api_buckets, state.Name.ValueString()), payload.S3ObjectLock, 200)
	if err != nil {
		return nil, fmt.Errorf("unable to update StorageGrid container: %w", err)
	}

	updatedObjectLockConfiguration, err := c.readObjectLockConfiguration(ctx, state.Name.ValueString())
	if err != nil {
		return nil, fmt.Errorf("unable to read updated StorageGrid container: %w", err)
	}

	return &BucketResourceModel{Name: state.Name, Region: state.Region, ObjectLockConfiguration: updatedObjectLockConfiguration}, nil
}

// Delete deletes the StorageGrid bucket with the given name.
func (c *BucketClient) Delete(bucketName string) error {
	if _, _, _, err := c.apiClient.SendRequest("DELETE", api_buckets+"/"+bucketName, nil, 204); err != nil {
		return fmt.Errorf("unable to delete StorageGrid container: %w", err)
	}
	return nil
}

// Read reads the StorageGrid bucket with the given name and returns the corresponding BucketResourceModel.
func (c *BucketClient) Read(ctx context.Context, bucketName string) (*BucketResourceModel, error) {
	type regionResult struct {
		val *string
		err error
	}
	type olcResult struct {
		val *ObjectLockConfiguration
		err error
	}

	regCh := make(chan regionResult, 1)
	olcCh := make(chan olcResult, 1)

	go func() {
		v, err := c.readRegion(ctx, bucketName)
		regCh <- regionResult{val: v, err: err}
	}()

	go func() {
		v, err := c.readObjectLockConfiguration(ctx, bucketName)
		olcCh <- olcResult{val: v, err: err}
	}()

	reg := <-regCh
	olc := <-olcCh

	combinedErrs := errors.Join(reg.err, olc.err)
	if combinedErrs != nil {
		return nil, combinedErrs
	}

	return &BucketResourceModel{
		Name:                    types.StringValue(bucketName),
		Region:                  types.StringValue(*reg.val),
		ObjectLockConfiguration: olc.val,
	}, nil
}

func (c *BucketClient) readRegion(ctx context.Context, bucketName string) (*string, error) {
	tflog.Debug(ctx, "1. Get refreshed bucket information.")
	endpoint := fmt.Sprintf("%s/%s/region", api_buckets, bucketName)
	respBody, _, respCode, err := c.apiClient.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			return nil, ErrBucketNotFound
		}
		return nil, &GenericError{Summary: "Error Reading StorageGrid container", Details: "Could not read StorageGrid container name " + bucketName + ": " + err.Error()}
	}

	type regionDataModel struct {
		Region string `json:"region"`
	}

	type regionReadModel struct {
		Data regionDataModel `json:"data"`
	}

	var returnBody regionReadModel

	tflog.Debug(ctx, "2. Unmarshal bucket information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse region response, got error: " + err.Error()}
	}

	return &returnBody.Data.Region, nil
}

func (c *BucketClient) readObjectLockConfiguration(ctx context.Context, bucketName string) (*ObjectLockConfiguration, error) {
	tflog.Debug(ctx, "1. Get refreshed bucket information.")
	endpoint := fmt.Sprintf("%s/%s/object-lock", api_buckets, bucketName)
	respBody, _, respCode, err := c.apiClient.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			return nil, ErrBucketNotFound
		}
		return nil, fmt.Errorf("unable to read object lock configuration: %w", err)
	}

	type retentionSettings struct {
		Mode  string  `json:"mode"`
		Days  *string `json:"days,omitempty"`  // (!) actually documented as int, but API returns values as strings
		Years *string `json:"years,omitempty"` // (!) actually documented as int, but API returns values as strings
	}

	type s3ObjectLock struct {
		Enabled           bool              `json:"enabled"`
		RetentionSettings retentionSettings `json:"defaultRetentionSetting"`
	}

	type objectLockConfigurationReadModel struct {
		Data s3ObjectLock `json:"data"`
	}

	var returnBody objectLockConfigurationReadModel

	tflog.Debug(ctx, "2. Unmarshal bucket information to JSON body.")
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse object lock configuration response, got error: " + err.Error()}
	}

	if !returnBody.Data.Enabled {
		return nil, nil
	}

	objectLockConfiguration := ObjectLockConfiguration{
		Mode: types.StringValue(returnBody.Data.RetentionSettings.Mode),
	}

	if strDays := returnBody.Data.RetentionSettings.Days; strDays != nil && *strDays != "" {
		days, err := strconv.Atoi(*strDays)
		if err != nil {
			return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse object lock configuration's retention days, got error: " + err.Error()}
		}
		objectLockConfiguration.Days = types.Int64Value(int64(days))
	}

	if strYears := returnBody.Data.RetentionSettings.Years; strYears != nil && *strYears != "" {
		years, err := strconv.Atoi(*strYears)
		if err != nil {
			return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse object lock configuration's retention years, got error: " + err.Error()}
		}
		objectLockConfiguration.Years = types.Int64Value(int64(years))
	}

	return &objectLockConfiguration, nil
}
