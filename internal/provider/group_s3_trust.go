// Copyright github.com/dmpe 2024, 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"encoding/json"

	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

const (
	s3TrustDescription          = "Trust policy that specifies the principals allowed to assume this group through the StorageGRID AssumeRole API."
	s3TrustResourceDescription  = s3TrustDescription + " Omitting this block removes any trust policy that was set outside Terraform."
	s3TrustPrincipalDescription = "ARNs/URNs of the principals (Principal.AWS) allowed to assume this group, e.g. urn:sgws:identity::<account_id>:user/<name>."
)

type S3TrustPolicyDataModel struct {
	Version   types.String                 `tfsdk:"version"`
	Statement []*S3TrustStatementDataModel `tfsdk:"statement"`
}

type S3TrustStatementDataModel struct {
	Sid          types.String   `tfsdk:"sid"`
	Effect       types.String   `tfsdk:"effect"`
	Action       []types.String `tfsdk:"action"`
	PrincipalAWS []types.String `tfsdk:"principal_aws"`
}

type S3TrustPolicy struct {
	Version   string             `json:"Version,omitempty"`
	Statement []S3TrustStatement `json:"Statement"`
}

type S3TrustStatement struct {
	Sid       string           `json:"Sid,omitempty"`
	Effect    string           `json:"Effect"`
	Action    stringOrList     `json:"Action"`
	Principal S3TrustPrincipal `json:"Principal"`
}

type S3TrustPrincipal struct {
	AWS stringOrList `json:"AWS"`
}

// Groups created outside Terraform may carry a bare string principal; fall back to that form so
// one such group does not break reading all groups.
func (p *S3TrustPrincipal) UnmarshalJSON(data []byte) error {
	var object struct {
		AWS stringOrList `json:"AWS"`
	}
	if err := json.Unmarshal(data, &object); err == nil {
		p.AWS = object.AWS
		return nil
	}
	return json.Unmarshal(data, &p.AWS)
}

// stringOrList accepts both forms the API uses and writes a single value as a plain string,
// which is how the API documents Principal.AWS.
type stringOrList []string

func (s *stringOrList) UnmarshalJSON(data []byte) error {
	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = stringOrList{single}
		return nil
	}
	var list []string
	if err := json.Unmarshal(data, &list); err != nil {
		return err
	}
	*s = list
	return nil
}

func (s stringOrList) MarshalJSON() ([]byte, error) {
	if len(s) == 1 {
		return json.Marshal(s[0])
	}
	return json.Marshal([]string(s))
}

func s3TrustPolicyToAPI(model *S3TrustPolicyDataModel) *S3TrustPolicy {
	if model == nil {
		return nil
	}

	policy := &S3TrustPolicy{
		Version:   model.Version.ValueString(),
		Statement: []S3TrustStatement{},
	}
	for _, st := range model.Statement {
		policy.Statement = append(policy.Statement, S3TrustStatement{
			Sid:       st.Sid.ValueString(),
			Effect:    st.Effect.ValueString(),
			Action:    toJson(st.Action),
			Principal: S3TrustPrincipal{AWS: toJson(st.PrincipalAWS)},
		})
	}
	return policy
}

func s3TrustPolicyFromAPI(policy *S3TrustPolicy) *S3TrustPolicyDataModel {
	if policy == nil || len(policy.Statement) == 0 {
		return nil
	}

	model := &S3TrustPolicyDataModel{
		Version: optionalString(policy.Version),
	}
	for _, st := range policy.Statement {
		model.Statement = append(model.Statement, &S3TrustStatementDataModel{
			Sid:          optionalString(st.Sid),
			Effect:       types.StringValue(st.Effect),
			Action:       toTerraform(st.Action),
			PrincipalAWS: toTerraform(st.Principal.AWS),
		})
	}
	return model
}

func optionalString(value string) types.String {
	if value == "" {
		return types.StringNull()
	}
	return types.StringValue(value)
}

func s3TrustResourceSchema() schema.SingleNestedAttribute {
	return schema.SingleNestedAttribute{
		Optional:    true,
		Description: s3TrustResourceDescription,
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Policy language version, e.g. 2012-10-17.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"statement": schema.ListNestedAttribute{
				Required:    true,
				Description: "a list of trust policy statements",
				Validators: []validator.List{
					listvalidator.SizeAtLeast(1),
				},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"sid": schema.StringAttribute{
							Optional:    true,
							Description: "an optional identifier that you provide for the policy statement",
						},
						"effect": schema.StringAttribute{
							Required:    true,
							Description: "the specific result of the statement (either an allow or an explicit deny)",
							Validators: []validator.String{
								stringvalidator.OneOf("Allow", "Deny"),
							},
						},
						act: schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: "the actions this statement applies to, e.g. sts:AssumeRole",
						},
						"principal_aws": schema.ListAttribute{
							ElementType: types.StringType,
							Required:    true,
							Description: s3TrustPrincipalDescription,
							Validators: []validator.List{
								listvalidator.SizeAtLeast(1),
							},
						},
					},
				},
			},
		},
	}
}

func s3TrustDataSourceSchema() dsschema.SingleNestedAttribute {
	return dsschema.SingleNestedAttribute{
		Computed:    true,
		Description: s3TrustDescription,
		Attributes: map[string]dsschema.Attribute{
			"version": dsschema.StringAttribute{
				Computed:    true,
				Description: "Policy language version.",
			},
			"statement": dsschema.ListNestedAttribute{
				Computed:    true,
				Description: "a list of trust policy statements",
				NestedObject: dsschema.NestedAttributeObject{
					Attributes: map[string]dsschema.Attribute{
						"sid": dsschema.StringAttribute{
							Computed:    true,
							Description: "an optional identifier that you provide for the policy statement",
						},
						"effect": dsschema.StringAttribute{
							Computed:    true,
							Description: "the specific result of the statement (either an allow or an explicit deny)",
						},
						act: dsschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: "the actions this statement applies to",
						},
						"principal_aws": dsschema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
							Description: s3TrustPrincipalDescription,
						},
					},
				},
			},
		},
	}
}
