// Copyright (c) github.com/dmpe
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type StatementResourceModel struct {
	Sid          types.String            `tfsdk:"sid"`
	Effect       types.String            `tfsdk:"effect"`
	Action       []types.String          `tfsdk:"action"`
	NotAction    []types.String          `tfsdk:"not_action"`
	Resource     []types.String          `tfsdk:"resource"`
	NotResource  []types.String          `tfsdk:"not_resource"`
	Condition    types.Map               `tfsdk:"condition"`
	Principal    *PrincipalResourceModel `tfsdk:"principal"`
	NotPrincipal *PrincipalResourceModel `tfsdk:"not_principal"`
}

type PolicyResourceModel struct {
	Id        types.String             `tfsdk:"id"`
	Version   types.String             `tfsdk:"version"`
	Statement []StatementResourceModel `tfsdk:"statement"`
}

type BucketPolicyResourceModel struct {
	BucketName types.String         `tfsdk:"bucket_name"`
	Policy     *PolicyResourceModel `tfsdk:"policy"`
}

type PrincipalResourceModel struct {
	Type        types.String   `tfsdk:"type"`
	Identifiers []types.String `tfsdk:"identifiers"`
}

func (m *PrincipalResourceModel) toPrincipalApiModel() any {
	if m.Type.ValueString() == "*" {
		return "*"
	}

	var identifiers any
	if m.Type.ValueString() == "AWS" {
		switch len(m.Identifiers) {
		case 0:
			identifiers = "*"
		case 1:
			identifiers = m.Identifiers[0].ValueString()
		default:
			ids := make([]string, len(m.Identifiers))
			for i, identifier := range m.Identifiers {
				ids[i] = identifier.ValueString()
			}
			identifiers = ids
		}
	}

	return map[string]any{"AWS": identifiers}
}

func NewPrincipalResourceModel(input any) (*PrincipalResourceModel, error) {
	if input == nil {
		return nil, nil
	}

	if _, ok := input.(string); ok {
		return &PrincipalResourceModel{
			Type:        types.StringValue("*"),
			Identifiers: nil,
		}, nil
	}

	if principal, ok := input.(map[string]any); ok {
		if _, ok := principal["AWS"]; !ok {
			return nil, fmt.Errorf("no 'AWS' key in principal type")
		}
		identifiers := principal["AWS"]
		if identifiers == "*" {
			return &PrincipalResourceModel{
				Type:        types.StringValue("AWS"),
				Identifiers: nil,
			}, nil
		}
		if singleIdentifier, ok := identifiers.(string); ok {
			return &PrincipalResourceModel{
				Type:        types.StringValue("AWS"),
				Identifiers: []types.String{types.StringValue(singleIdentifier)},
			}, nil
		}
		if multipleIdentifiers, ok := identifiers.([]interface{}); ok {
			identifiersTf := make([]types.String, len(multipleIdentifiers))
			for i, identifier := range multipleIdentifiers {
				assertedIdentifier, ok := identifier.(string)
				if !ok {
					return nil, fmt.Errorf("invalid identifier type. expected string, got %T", identifier)
				}
				identifiersTf[i] = types.StringValue(assertedIdentifier)
			}
			return &PrincipalResourceModel{
				Type:        types.StringValue("AWS"),
				Identifiers: identifiersTf,
			}, nil
		}
	}
	return nil, fmt.Errorf("invalid principal type")
}

func (m *BucketPolicyResourceModel) toBucketPolicyApiModel(ctx context.Context, diagnostics *diag.Diagnostics) *BucketPolicyApiModel {
	statement := make([]StatementApiModel, len(m.Policy.Statement))
	for i, stmt := range m.Policy.Statement {
		var principal any
		if stmt.Principal != nil {
			principal = stmt.Principal.toPrincipalApiModel()
		}

		var notPrincipal any
		if stmt.NotPrincipal != nil {
			notPrincipal = stmt.NotPrincipal.toPrincipalApiModel()
		}

		conditionOperators := make(map[string]types.Map, len(m.Policy.Statement[i].Condition.Elements()))
		if diags := m.Policy.Statement[i].Condition.ElementsAs(ctx, &conditionOperators, false); diags.HasError() {
			diagnostics.Append(diags...)
			return nil
		}

		condition := make(map[string]map[string]string, len(conditionOperators))

		for operator, operatorValue := range conditionOperators {
			conditionKeys := make(map[string]types.String, len(operatorValue.Elements()))
			if diags := operatorValue.ElementsAs(ctx, &conditionKeys, false); diags.HasError() {
				diagnostics.Append(diags...)
				return nil
			}
			keyMap := make(map[string]string, len(conditionKeys))
			for key, conditionValue := range conditionKeys {
				keyMap[key] = conditionValue.ValueString()
			}
			condition[operator] = keyMap
		}

		statement[i] = StatementApiModel{
			Sid:          stmt.Sid.ValueString(),
			Effect:       stmt.Effect.ValueString(),
			Action:       toJson(stmt.Action),
			NotAction:    toJson(stmt.NotAction),
			Resource:     toJson(stmt.Resource),
			NotResource:  toJson(stmt.NotResource),
			Condition:    condition,
			Principal:    principal,
			NotPrincipal: notPrincipal,
		}
	}

	return &BucketPolicyApiModel{
		Policy: &PolicyApiModel{
			Id:        m.Policy.Id.ValueString(),
			Version:   m.Policy.Version.ValueString(),
			Statement: statement,
		},
	}
}

func (m *BucketPolicyResourceModel) upsert(ctx context.Context, client HttpClient, diagnostics *diag.Diagnostics) *BucketPolicyResourceModel {
	endpoint := fmt.Sprintf("%s/%s/policy", api_buckets, m.BucketName.ValueString())

	payload := m.toBucketPolicyApiModel(ctx, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	respBody, _, respCode, err := client.SendRequest("PUT", endpoint, payload, 200)
	if err != nil {
		if respCode == http.StatusBadRequest {
			diagnostics.AddError("invalid bucket policy", fmt.Sprintf("invalid bucket policy for bucket '%s': %s", m.BucketName.ValueString(), err.Error()))
			return nil
		}
		if respCode == http.StatusNotFound {
			diagnostics.AddError("bucket not found", fmt.Sprintf("bucket '%s' not found", m.BucketName.ValueString()))
			return nil
		}
		diagnostics.AddError("unable to create or update bucket policy", err.Error())
		return nil
	}

	return NewBucketPolicyResourceModel(m.BucketName.ValueString(), respBody, diagnostics)
}

func (m *BucketPolicyResourceModel) read(client HttpClient, diagnostics *diag.Diagnostics) *BucketPolicyResourceModel {
	endpoint := fmt.Sprintf("%s/%s/policy", api_buckets, m.BucketName.ValueString())
	respBody, _, respCode, err := client.SendRequest("GET", endpoint, nil, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			diagnostics.AddError("bucket not found", fmt.Sprintf("bucket '%s' not found", m.BucketName.ValueString()))
			return nil
		}
		diagnostics.AddError("unable to read bucket policy", err.Error())
		return nil
	}

	return NewBucketPolicyResourceModel(m.BucketName.ValueString(), respBody, diagnostics)
}

func (m *BucketPolicyResourceModel) delete(client HttpClient) error {
	endpoint := fmt.Sprintf("%s/%s/policy", api_buckets, m.BucketName.ValueString())

	payload := BucketPolicyApiModel{Policy: nil}

	_, _, respCode, err := client.SendRequest("PUT", endpoint, payload, 200)
	if err != nil {
		if respCode == http.StatusNotFound {
			return ErrBucketNotFound
		}
		return fmt.Errorf("unable to delete bucket policy: %w", err)
	}

	return nil
}

// NewBucketPolicyResourceModel parses the JSON response from the API into a BucketPolicyResourceModel.
func NewBucketPolicyResourceModel(bucketName string, input []byte, diagnostics *diag.Diagnostics) *BucketPolicyResourceModel {
	type responseDataType struct {
		Data BucketPolicyApiModel `json:"data"`
	}

	var returnBody responseDataType
	if err := json.Unmarshal(input, &returnBody); err != nil {
		diagnostics.AddError("unable to parse bucket policy response", err.Error())
		return nil
	}

	statement := make([]StatementResourceModel, len(returnBody.Data.Policy.Statement))
	for i, stmt := range returnBody.Data.Policy.Statement {
		s := NewStatementResourceModel(stmt, diagnostics)
		if diagnostics.HasError() {
			return nil
		}
		if s == nil {
			diagnostics.AddError("unexpected nil value for policy statement", fmt.Sprintf("statement #%d", i))
			return nil
		}
		statement[i] = *s
	}

	return &BucketPolicyResourceModel{
		BucketName: types.StringValue(bucketName),
		Policy: &PolicyResourceModel{
			Id:        types.StringValue(returnBody.Data.Policy.Id),
			Version:   types.StringValue(returnBody.Data.Policy.Version),
			Statement: statement,
		},
	}
}

// NewStatementResourceModel parses the JSON response from the API into a StatementResourceModel.
func NewStatementResourceModel(input StatementApiModel, diagnostics *diag.Diagnostics) *StatementResourceModel {
	principal, err := NewPrincipalResourceModel(input.Principal)
	if err != nil {
		diagnostics.AddError("failed to create principal resource model", err.Error())
		return nil
	}

	nonPrincipal, err := NewPrincipalResourceModel(input.NotPrincipal)
	if err != nil {
		diagnostics.AddError("failed to create non-principal resource model", err.Error())
		return nil
	}

	condition := mapOfMapsToTerraform(input.Condition, diagnostics)
	if diagnostics.HasError() {
		return nil
	}

	return &StatementResourceModel{
		Sid:          types.StringValue(input.Sid),
		Effect:       types.StringValue(input.Effect),
		Action:       toTerraform(input.Action),
		NotAction:    toTerraform(input.NotAction),
		Resource:     toTerraform(input.Resource),
		NotResource:  toTerraform(input.NotResource),
		Condition:    condition,
		Principal:    principal,
		NotPrincipal: nonPrincipal,
	}
}
