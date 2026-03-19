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

func TestAuditLog_CreateAndDeleteRange(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	app := server.NewApp(database)

	domainID, _ := setupDomainAndParent(t, app)

	_, createBody := doRequest(app, "POST", "/api/v1/ranges", map[string]interface{}{
		"name":   "audited-range",
		"cidr":   "10.30.0.0/24",
		"domain": fmt.Sprintf("%d", domainID),
	})
	var created map[string]interface{}
	require.NoError(t, json.Unmarshal(createBody, &created))
	id := int(created["id"].(float64))

	doRequest(app, "DELETE", fmt.Sprintf("/api/v1/ranges/%d", id), nil)

	status, body := doRequestWithStatus(app, "GET", "/api/v1/audit", nil)
	assert.Equal(t, 200, status)

	var logs []map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &logs))

	assert.GreaterOrEqual(t, len(logs), 2)
	assert.Equal(t, "delete", logs[0]["action"])
	assert.Equal(t, "range", logs[0]["resource_type"])
	assert.Equal(t, "audited-range", logs[0]["resource_name"])
}

func TestAuditLog_CreateDomain(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	app := server.NewApp(database)

	doRequest(app, "POST", "/api/v1/domains", map[string]interface{}{
		"name": "audit-domain",
		"vpcs": []string{},
	})

	status, body := doRequestWithStatus(app, "GET", "/api/v1/audit", nil)
	assert.Equal(t, 200, status)

	var logs []map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &logs))
	require.NotEmpty(t, logs)
	assert.Equal(t, "create", logs[0]["action"])
	assert.Equal(t, "domain", logs[0]["resource_type"])
	assert.Equal(t, "audit-domain", logs[0]["resource_name"])
}

func TestAuditLog_LimitParam(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	app := server.NewApp(database)

	domainID, _ := setupDomainAndParent(t, app)
	for i := range 5 {
		doRequest(app, "POST", "/api/v1/ranges", map[string]interface{}{
			"name":   fmt.Sprintf("range-%d", i),
			"cidr":   fmt.Sprintf("10.40.%d.0/24", i),
			"domain": fmt.Sprintf("%d", domainID),
		})
	}

	status, body := doRequestWithStatus(app, "GET", "/api/v1/audit?limit=2", nil)
	assert.Equal(t, 200, status)

	var logs []map[string]interface{}
	require.NoError(t, json.Unmarshal(body, &logs))
	assert.Len(t, logs, 2)
}
