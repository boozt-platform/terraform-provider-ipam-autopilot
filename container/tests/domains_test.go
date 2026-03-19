// Copyright 2026 Boozt Fashion AB
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:build integration

package tests

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/boozt-platform/ipam-autopilot/container/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateDomain(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	app := server.NewApp(database)

	status, body := doRequestWithStatus(app, "POST", "/api/v1/domains", map[string]interface{}{
		"name": "test-domain",
		"vpcs": []string{"projects/test/networks/vpc1"},
	})
	assert.Equal(t, 200, status)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.NotNil(t, resp["id"])
}

func TestGetDomains(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	app := server.NewApp(database)

	doRequest(app, "POST", "/api/v1/domains", map[string]interface{}{
		"name": "test-domain",
		"vpcs": []string{},
	})

	status, body := doRequestWithStatus(app, "GET", "/api/v1/domains", nil)
	assert.Equal(t, 200, status)

	var resp []interface{}
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Len(t, resp, 1)
}

func TestGetDomain(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	app := server.NewApp(database)

	_, createBody := doRequest(app, "POST", "/api/v1/domains", map[string]interface{}{
		"name": "test-domain",
		"vpcs": []string{},
	})
	var created map[string]interface{}
	require.NoError(t, json.Unmarshal(createBody, &created))
	id := int(created["id"].(float64))

	status, body := doRequestWithStatus(app, "GET", fmt.Sprintf("/api/v1/domains/%d", id), nil)
	assert.Equal(t, 200, status)

	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &resp))
	assert.Equal(t, "test-domain", resp["name"])
}

func TestUpdateDomain(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	app := server.NewApp(database)

	_, createBody := doRequest(app, "POST", "/api/v1/domains", map[string]interface{}{
		"name": "original",
		"vpcs": []string{},
	})
	var created map[string]interface{}
	require.NoError(t, json.Unmarshal(createBody, &created))
	id := int(created["id"].(float64))

	status, _ := doRequestWithStatus(app, "PUT", fmt.Sprintf("/api/v1/domains/%d", id), map[string]interface{}{
		"name": "updated",
	})
	assert.Equal(t, 200, status)

	_, getBody := doRequest(app, "GET", fmt.Sprintf("/api/v1/domains/%d", id), nil)
	var resp map[string]interface{}
	require.NoError(t, json.Unmarshal(getBody, &resp))
	assert.Equal(t, "updated", resp["name"])
}

func TestDeleteDomain(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	app := server.NewApp(database)

	_, createBody := doRequest(app, "POST", "/api/v1/domains", map[string]interface{}{
		"name": "to-delete",
		"vpcs": []string{},
	})
	var created map[string]interface{}
	require.NoError(t, json.Unmarshal(createBody, &created))
	id := int(created["id"].(float64))

	status, _ := doRequestWithStatus(app, "DELETE", fmt.Sprintf("/api/v1/domains/%d", id), nil)
	assert.Equal(t, 200, status)
}
