// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BucketVersioningResourceModel struct {
	BucketName types.String `tfsdk:"bucket_name"`
	Status     types.String `tfsdk:"status"`
}

func (m *BucketVersioningResourceModel) ToBucketVersioningApiRequestModel() BucketVersioningApiRequestModel {
	return BucketVersioningApiRequestModel{
		IsEnabled:   m.Status.ValueString() == "Enabled",
		IsSuspended: m.Status.ValueString() == "Suspended",
	}
}

func (m *BucketVersioningResourceModel) upsert(client HttpClient) (*BucketVersioningResourceModel, error) {
	endpoint := fmt.Sprintf("%s/%s/versioning", api_buckets, m.BucketName.ValueString())
	httpResp, _, _, err := client.SendRequest("PUT", endpoint, m.ToBucketVersioningApiRequestModel(), 200)
	if err != nil {
		return nil, fmt.Errorf("unable to create or update bucket versioning: %w", err)
	}

	var returnBody BucketVersioningApiResponseModel
	if err := json.Unmarshal(httpResp, &returnBody); err != nil {
		return nil, fmt.Errorf("unable to unmarshal create or update bucket versioning response: %w", err)
	}

	return &BucketVersioningResourceModel{
		BucketName: m.BucketName,
		Status:     types.StringValue(returnBody.Status()),
	}, nil
}

func (m *BucketVersioningResourceModel) read(client HttpClient) (*BucketVersioningResourceModel, error) {
	endpoint := fmt.Sprintf("%s/%s/versioning", api_buckets, m.BucketName.ValueString())
	respBody, _, _, err := client.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		return nil, fmt.Errorf("unable to read bucket versioning: %w", err)
	}

	var returnBody BucketVersioningApiResponseModel
	if err := json.Unmarshal(respBody, &returnBody); err != nil {
		return nil, fmt.Errorf("unable to parse bucket versioning read response: %w", err)
	}

	return &BucketVersioningResourceModel{
		BucketName: m.BucketName,
		Status:     types.StringValue(returnBody.Status()),
	}, nil
}

type BucketVersioningApiRequestModel struct {
	IsEnabled   bool `json:"versioningEnabled"`
	IsSuspended bool `json:"versioningSuspended"`
}

type BucketVersioningApiResponseModel struct {
	Data BucketVersioningApiRequestModel `json:"data"`
}

func (m *BucketVersioningApiResponseModel) Status() (status string) {
	if m.Data.IsSuspended {
		status = "Suspended"
	} else {
		if m.Data.IsEnabled {
			status = "Enabled"
		} else {
			status = "Disabled"
		}
	}
	return
}
