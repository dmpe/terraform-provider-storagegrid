// Copyright github.com/dmpe 2024, 2026
// SPDX-License-Identifier: MIT

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/stretchr/testify/assert"
)

// mockTenant emulates the subset of the StorageGRID Tenant Management API that
// the users resource uses: POST /api/v4/authorize and GET /api/v4/org/users/{id}.
type mockTenant struct {
	server *httptest.Server
	users  map[string]UserModel
}

func newMockTenant(t *testing.T, users map[string]UserModel) *mockTenant {
	t.Helper()
	m := &mockTenant{users: users}
	mux := http.NewServeMux()

	mux.HandleFunc("/api/v4/authorize", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(S3GridClientReturnJson{Data: "mock-token"})
	})

	mux.HandleFunc("/api/v4/org/users/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		user, ok := m.users[strings.TrimPrefix(r.URL.Path, "/api/v4/org/users/")]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			_ = json.NewEncoder(w).Encode(map[string]string{"message": "No such user"})
			return
		}
		_ = json.NewEncoder(w).Encode(UsersDataModelSingle{Data: user})
	})

	m.server = httptest.NewServer(mux)
	t.Cleanup(m.server.Close)
	return m
}

const (
	mockUserID     = "11111111-1111-1111-1111-111111111111"
	mockGroupID    = "3aedfab3-ea53-441b-b8f7-8d3cb817bb9f"
	mockTenantID   = "9167766278272721"
	mockUniqueName = "user/mock-user"
)

func mockUserModel(memberOf []string) UserModel {
	return UserModel{
		UniqueName: mockUniqueName,
		FullName:   "Mock User",
		MemberOf:   memberOf,
		Disable:    false,
		AccountId:  mockTenantID,
		ID:         mockUserID,
		Federated:  false,
		UserURN:    "urn:esg:tenant:" + mockTenantID + ":user:mock-user",
	}
}

func newUsersResourceForTest(baseURL string) *usersResource {
	r := &usersResource{}
	r.client = NewTokenClient(baseURL, "mock-token", false)
	return r
}

// newUsersReadRequest builds a Read request whose state mirrors what the plugin
// framework produces during import: only the passthrough id is set, every other
// attribute is null. Pass memberOf to simulate a refresh against prior state.
func newUsersReadRequest(t *testing.T, r *usersResource, memberOf []string) (resource.ReadRequest, *resource.ReadResponse) {
	t.Helper()
	ctx := context.Background()

	var schemaResp resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
	objType, ok := schemaResp.Schema.Type().TerraformType(ctx).(tftypes.Object)
	if !ok {
		t.Fatal("users resource schema type is not an object")
	}

	raw := tftypes.NewValue(objType, map[string]tftypes.Value{
		"unique_name": tftypes.NewValue(tftypes.String, nil),
		"full_name":   tftypes.NewValue(tftypes.String, nil),
		"disable":     tftypes.NewValue(tftypes.Bool, nil),
		"member_of":   mockMemberOfValue(t, objType, memberOf),
		"federated":   tftypes.NewValue(tftypes.Bool, nil),
		"user_urn":    tftypes.NewValue(tftypes.String, nil),
		"account_id":  tftypes.NewValue(tftypes.String, nil),
		"id":          tftypes.NewValue(tftypes.String, mockUserID),
	})

	req := resource.ReadRequest{
		State: tfsdk.State{Schema: schemaResp.Schema, Raw: raw},
	}
	resp := &resource.ReadResponse{
		State: tfsdk.State{Schema: schemaResp.Schema},
	}
	return req, resp
}

func mockMemberOfValue(t *testing.T, objType tftypes.Object, values []string) tftypes.Value {
	t.Helper()
	listType, ok := objType.AttributeTypes["member_of"].(tftypes.List)
	if !ok {
		t.Fatal("member_of attribute type is not a list")
	}
	if values == nil {
		return tftypes.NewValue(listType, nil)
	}
	listValues := make([]tftypes.Value, 0, len(values))
	for _, v := range values {
		listValues = append(listValues, tftypes.NewValue(tftypes.String, v))
	}
	return tftypes.NewValue(listType, listValues)
}

func readUsersState(t *testing.T, resp *resource.ReadResponse) usersDataSourceDataModel {
	t.Helper()
	var state usersDataSourceDataModel
	diags := resp.State.Get(context.Background(), &state)
	assert.Empty(t, diags)
	return state
}

// Regression test for the import bug: during import the state contains only the
// passthrough id, so state member_of is empty while the API returns real
// memberships. Read must adopt the API value instead of failing.
func TestUsersResourceReadFromImportStateWithGroups(t *testing.T) {
	m := newMockTenant(t, map[string]UserModel{
		mockUserID: mockUserModel([]string{mockGroupID}),
	})
	r := newUsersResourceForTest(m.server.URL)
	ctx := context.Background()

	req, resp := newUsersReadRequest(t, r, nil)
	r.Read(ctx, req, resp)

	assert.Empty(t, resp.Diagnostics, "import of a grouped user must not produce diagnostics")

	state := readUsersState(t, resp)
	assert.Equal(t, mockUserID, state.ID.ValueString())
	assert.Equal(t, mockUniqueName, state.UniqueName.ValueString())
	assert.Equal(t, mockTenantID, state.AccountId.ValueString())
	assert.Equal(t, []types.String{types.StringValue(mockGroupID)}, state.MemberOf)
}

// Sets that are equal in any order keep the state order to avoid false diffs.
func TestUsersResourceReadKeepsStateOrderWhenSetsEqual(t *testing.T) {
	m := newMockTenant(t, map[string]UserModel{
		mockUserID: mockUserModel([]string{"group-2", "group-1"}),
	})
	r := newUsersResourceForTest(m.server.URL)
	ctx := context.Background()

	req, resp := newUsersReadRequest(t, r, []string{"group-1", "group-2"})
	r.Read(ctx, req, resp)

	assert.Empty(t, resp.Diagnostics)

	state := readUsersState(t, resp)
	assert.Equal(t, []types.String{
		types.StringValue("group-1"),
		types.StringValue("group-2"),
	}, state.MemberOf)
}

// An out-of-band membership change must be adopted as refreshed data, not
// turned into a hard error.
func TestUsersResourceReadAdoptsMembershipDrift(t *testing.T) {
	m := newMockTenant(t, map[string]UserModel{
		mockUserID: mockUserModel([]string{mockGroupID, "22222222-2222-2222-2222-222222222222"}),
	})
	r := newUsersResourceForTest(m.server.URL)
	ctx := context.Background()

	req, resp := newUsersReadRequest(t, r, []string{mockGroupID})
	r.Read(ctx, req, resp)

	assert.Empty(t, resp.Diagnostics)

	state := readUsersState(t, resp)
	assert.Equal(t, []types.String{
		types.StringValue(mockGroupID),
		types.StringValue("22222222-2222-2222-2222-222222222222"),
	}, state.MemberOf)
}

func TestUsersResourceReadNotFoundRemovesResource(t *testing.T) {
	m := newMockTenant(t, map[string]UserModel{})
	r := newUsersResourceForTest(m.server.URL)
	ctx := context.Background()

	req, resp := newUsersReadRequest(t, r, nil)
	r.Read(ctx, req, resp)

	assert.False(t, resp.Diagnostics.HasError())
	assert.True(t, resp.State.Raw.IsNull(), "a 404 must remove the resource from state")
}
