// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BucketQuotaResourceModel struct {
	BucketName  types.String `tfsdk:"bucket_name"`
	ObjectBytes types.Int64  `tfsdk:"object_bytes"`
}

type BucketQuotaApiModel struct {
	ObjectBytes *int `json:"quotaObjectBytes"`
}

func (m *BucketQuotaResourceModel) upsert(client HttpClient) (*BucketQuotaResourceModel, error) {
	endpoint := fmt.Sprintf("%s/%s/quota-object-bytes", api_buckets, m.BucketName.ValueString())

	objectBytes := int(m.ObjectBytes.ValueInt64())
	payload := BucketQuotaApiModel{ObjectBytes: &objectBytes}

	respBody, _, respCode, err := client.SendRequest("PUT", endpoint, payload, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			return nil, ErrBucketNotFound
		}
		return nil, fmt.Errorf("unable to create or update bucket quota: %w", err)
	}

	return NewBucketQuotaResourceModel(m.BucketName.ValueString(), respBody)
}

func (m *BucketQuotaResourceModel) read(client HttpClient) (*BucketQuotaResourceModel, error) {
	endpoint := fmt.Sprintf("%s/%s/quota-object-bytes", api_buckets, m.BucketName.ValueString())
	respBody, _, respCode, err := client.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			return nil, ErrBucketNotFound
		}
		return nil, fmt.Errorf("unable to read object lock configuration: %w", err)
	}

	return NewBucketQuotaResourceModel(m.BucketName.ValueString(), respBody)
}

func (m *BucketQuotaResourceModel) delete(client HttpClient) error {
	endpoint := fmt.Sprintf("%s/%s/quota-object-bytes", api_buckets, m.BucketName.ValueString())

	payload := BucketQuotaApiModel{ObjectBytes: nil}

	_, _, respCode, err := client.SendRequest("PUT", endpoint, payload, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			return ErrBucketNotFound
		}
		return fmt.Errorf("unable to delete object lock configuration: %w", err)
	}

	return nil
}

// NewBucketQuotaResourceModel parses the JSON response from the API into a BucketQuotaResourceModel.
func NewBucketQuotaResourceModel(bucketName string, input []byte) (*BucketQuotaResourceModel, error) {
	type responseDataType struct {
		Data BucketQuotaApiModel `json:"data"`
	}

	var returnBody responseDataType
	if err := json.Unmarshal(input, &returnBody); err != nil {
		return nil, &GenericError{Summary: "Client Error", Details: "Unable to parse object lock configuration response, got error: " + err.Error()}
	}

	var objectBytes *int64
	if returnBody.Data.ObjectBytes != nil {
		int64ObjectBytes := int64(*returnBody.Data.ObjectBytes)
		objectBytes = &int64ObjectBytes
	}

	return &BucketQuotaResourceModel{
		BucketName:  types.StringValue(bucketName),
		ObjectBytes: types.Int64PointerValue(objectBytes),
	}, nil
}
