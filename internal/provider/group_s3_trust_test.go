// Copyright github.com/dmpe 2024, 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"encoding/json"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

const testTrustPrincipal = "urn:sgws:identity::12345678901234567890:user/vault"

func TestS3TrustPolicyToAPIMarshalsSinglePrincipalAsString(t *testing.T) {
	model := &S3TrustPolicyDataModel{
		Version: types.StringValue("2012-10-17"),
		Statement: []*S3TrustStatementDataModel{{
			Sid:          types.StringNull(),
			Effect:       types.StringValue("Allow"),
			Action:       []types.String{types.StringValue("sts:AssumeRole")},
			PrincipalAWS: []types.String{types.StringValue(testTrustPrincipal)},
		}},
	}

	got, err := json.Marshal(GroupPostPolicies{S3Trust: s3TrustPolicyToAPI(model)})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded struct {
		S3Trust struct {
			Statement []struct {
				Sid       *string `json:"Sid"`
				Principal struct {
					AWS any `json:"AWS"`
				} `json:"Principal"`
			} `json:"Statement"`
		} `json:"s3Trust"`
	}
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if len(decoded.S3Trust.Statement) != 1 {
		t.Fatalf("expected 1 statement, got %s", got)
	}
	statement := decoded.S3Trust.Statement[0]
	if statement.Principal.AWS != testTrustPrincipal {
		t.Fatalf("Principal.AWS mismatch: got %v, want %q", statement.Principal.AWS, testTrustPrincipal)
	}
	if statement.Sid != nil {
		t.Fatalf("empty Sid should be omitted, got %s", got)
	}
}

func TestGroupPostPoliciesOmitsS3TrustWhenUnset(t *testing.T) {
	got, err := json.Marshal(GroupPostPolicies{S3Trust: s3TrustPolicyToAPI(nil)})
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var decoded map[string]any
	if err := json.Unmarshal(got, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if _, ok := decoded["s3Trust"]; ok {
		t.Fatalf("s3Trust should be omitted when unset, got %s", got)
	}
}

func TestS3TrustPolicyFromAPIAcceptsStringAndList(t *testing.T) {
	body := `{"s3Trust":{"Version":"2012-10-17","Statement":[
		{"Effect":"Allow","Action":"sts:AssumeRole","Principal":{"AWS":"` + testTrustPrincipal + `"}},
		{"Sid":"two","Effect":"Deny","Action":["sts:AssumeRole"],"Principal":{"AWS":["a","b"]}}
	]}}`

	var policies Policies
	if err := json.Unmarshal([]byte(body), &policies); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	model := s3TrustPolicyFromAPI(policies.S3Trust)
	if model == nil || len(model.Statement) != 2 {
		t.Fatalf("expected 2 statements, got %+v", model)
	}
	first, second := model.Statement[0], model.Statement[1]
	if !first.Sid.IsNull() || first.PrincipalAWS[0].ValueString() != testTrustPrincipal || first.Action[0].ValueString() != "sts:AssumeRole" {
		t.Fatalf("first statement mismatch: %+v", first)
	}
	if second.Sid.ValueString() != "two" || len(second.PrincipalAWS) != 2 {
		t.Fatalf("second statement mismatch: %+v", second)
	}
}

func TestS3TrustPrincipalAcceptsBareString(t *testing.T) {
	body := `{"s3Trust":{"Statement":[{"Effect":"Allow","Action":"sts:AssumeRole","Principal":"` + testTrustPrincipal + `"}]}}`

	var policies Policies
	if err := json.Unmarshal([]byte(body), &policies); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}

	model := s3TrustPolicyFromAPI(policies.S3Trust)
	if model == nil || len(model.Statement) != 1 || model.Statement[0].PrincipalAWS[0].ValueString() != testTrustPrincipal {
		t.Fatalf("bare string principal not accepted: %+v", model)
	}
}

func TestS3TrustPolicyFromAPIReturnsNilWithoutStatements(t *testing.T) {
	if got := s3TrustPolicyFromAPI(nil); got != nil {
		t.Fatalf("expected nil for missing policy, got %+v", got)
	}
	if got := s3TrustPolicyFromAPI(&S3TrustPolicy{}); got != nil {
		t.Fatalf("expected nil for empty policy, got %+v", got)
	}
}
