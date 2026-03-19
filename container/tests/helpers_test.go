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
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/boozt-platform/ipam-autopilot/container/server"
	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/require"
	tcmysql "github.com/testcontainers/testcontainers-go/modules/mysql"
)

func setupTestDB(t *testing.T) (*sql.DB, func()) {
	t.Helper()
	ctx := context.Background()

	container, err := tcmysql.Run(ctx,
		"mysql:8.4",
		tcmysql.WithDatabase("ipam"),
		tcmysql.WithUsername("ipam"),
		tcmysql.WithPassword("ipam"),
	)
	require.NoError(t, err)

	host, err := container.Host(ctx)
	require.NoError(t, err)
	port, err := container.MappedPort(ctx, "3306")
	require.NoError(t, err)

	cfg := mysql.Config{
		User:                 "ipam",
		Passwd:               "ipam",
		Net:                  "tcp",
		Addr:                 fmt.Sprintf("%s:%s", host, port.Port()),
		DBName:               "ipam",
		MultiStatements:      true,
		AllowNativePasswords: true,
	}

	database, err := sql.Open("mysql", cfg.FormatDSN())
	require.NoError(t, err)

	err = server.MigrateDatabase("ipam", database)
	require.NoError(t, err)

	cleanup := func() {
		database.Close()
		container.Terminate(ctx)
	}
	return database, cleanup
}

func doRequest(app *fiber.App, method, path string, body interface{}) (*httptest.ResponseRecorder, []byte) {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, _ := app.Test(req, 10000)
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return nil, respBody
}

func doRequestWithStatus(app *fiber.App, method, path string, body interface{}) (int, []byte) {
	var reqBody io.Reader
	if body != nil {
		b, _ := json.Marshal(body)
		reqBody = bytes.NewReader(b)
	}
	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, _ := app.Test(req, 10000)
	defer resp.Body.Close()
	respBody, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, respBody
}

func setupDomainAndParent(t *testing.T, app *fiber.App) (domainID int, parentID int) {
	t.Helper()

	_, domainBody := doRequest(app, "POST", "/api/v1/domains", map[string]interface{}{
		"name": "test-domain",
		"vpcs": []string{},
	})
	var domain map[string]interface{}
	require.NoError(t, json.Unmarshal(domainBody, &domain))
	domainID = int(domain["id"].(float64))

	_, parentBody := doRequest(app, "POST", "/api/v1/ranges", map[string]interface{}{
		"name":   "parent",
		"cidr":   "10.0.0.0/8",
		"domain": fmt.Sprintf("%d", domainID),
	})
	var parent map[string]interface{}
	require.NoError(t, json.Unmarshal(parentBody, &parent))
	parentID = int(parent["id"].(float64))

	return domainID, parentID
}
