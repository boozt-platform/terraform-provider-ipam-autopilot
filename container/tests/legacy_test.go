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
	"testing"

	"github.com/boozt-platform/ipam-autopilot/container/server"
	"github.com/stretchr/testify/assert"
)

func TestLegacyRoutesStillWork(t *testing.T) {
	database, cleanup := setupTestDB(t)
	defer cleanup()
	app := server.NewApp(database)

	// Old /domains path should still work (for Terraform provider compat)
	status, _ := doRequestWithStatus(app, "GET", "/domains", nil)
	assert.Equal(t, 200, status)

	status, _ = doRequestWithStatus(app, "GET", "/ranges", nil)
	assert.Equal(t, 200, status)
}
