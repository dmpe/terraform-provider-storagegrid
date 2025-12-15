// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type BucketPolicyApiModel struct {
	Policy *PolicyApiModel `json:"policy"`
}

type PolicyApiModel struct {
	Id        string              `json:"Id"`
	Version   string              `json:"Version"`
	Statement []StatementApiModel `json:"Statement"`
}

type StatementApiModel struct {
	Sid          string                       `json:"Sid"`
	Effect       string                       `json:"Effect"`
	Action       StringOrStrings              `json:"Action,omitempty"`
	NotAction    StringOrStrings              `json:"NotAction,omitempty"`
	Resource     StringOrStrings              `json:"Resource,omitempty"`
	NotResource  StringOrStrings              `json:"NotResource,omitempty"`
	Condition    map[string]map[string]string `json:"Condition,omitempty"`
	Principal    any                          `json:"Principal,omitempty"`
	NotPrincipal any                          `json:"NotPrincipal,omitempty"`
}

func toJson(in []types.String) []string {
	out := make([]string, len(in))
	for i, v := range in {
		out[i] = v.ValueString()
	}
	return out
}

func toTerraform(in []string) []types.String {
	out := make([]types.String, len(in))
	for i, v := range in {
		out[i] = types.StringValue(v)
	}
	return out
}

func mapOfMapsToTerraform(in map[string]map[string]string, diagnostics *diag.Diagnostics) types.Map {
	if in == nil {
		return types.MapNull(types.MapType{}.WithElementType(types.StringType))
	}

	var condition = map[string]attr.Value{}

	for operator, operatorValue := range in {
		inner := map[string]attr.Value{}
		for key, value := range operatorValue {
			inner[key] = types.StringValue(value)
		}
		innerValue, diags := types.MapValue(types.StringType, inner)
		if diags.HasError() {
			diagnostics.Append(diags...)
			return types.MapNull(types.MapType{}.WithElementType(types.StringType))
		}
		condition[operator] = innerValue
	}

	conditionValue, diags := types.MapValue(types.MapType{}.WithElementType(types.StringType), condition)
	if diags.HasError() {
		diagnostics.Append(diags...)
		return types.MapNull(types.MapType{}.WithElementType(types.StringType))
	}
	return conditionValue
}

type StringOrStrings []string

func (s *StringOrStrings) UnmarshalJSON(data []byte) error {
	data = bytes.TrimSpace(data)

	if bytes.Equal(data, []byte("null")) {
		*s = nil
		return nil
	}

	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = []string{single}
		return nil
	}

	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		*s = many
		return nil
	}

	return fmt.Errorf("expected string or []string, got: %s", string(data))
}
