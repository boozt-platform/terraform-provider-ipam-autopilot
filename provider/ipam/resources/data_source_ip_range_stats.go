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

package resources

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/boozt-platform/ipam-autopilot/provider/ipam/config"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func DataSourceIpRangeStats() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceStatsRead,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Name of the IP range to look up.",
			},
			"cidr": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "The CIDR block of the range.",
			},
			"total_addresses": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Total number of IP addresses in the range.",
			},
			"used_addresses": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of IP addresses allocated to direct child ranges.",
			},
			"free_addresses": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Number of IP addresses not yet allocated.",
			},
			"utilization_pct": {
				Type:        schema.TypeFloat,
				Computed:    true,
				Description: "Percentage of address space allocated to direct child ranges (0–100).",
			},
		},
	}
}

func dataSourceStatsRead(d *schema.ResourceData, meta interface{}) error {
	cfg := meta.(config.Config)
	name := d.Get("name").(string)

	accessToken, err := getIdentityToken(cfg.Url)
	if err != nil {
		return fmt.Errorf("unable to retrieve access token: %v", err)
	}

	// Resolve ID by name.
	listURL := fmt.Sprintf("%s/ranges?name=%s", cfg.Url, name)
	listReq, err := http.NewRequest("GET", listURL, nil)
	if err != nil {
		return fmt.Errorf("failed creating request: %v", err)
	}
	listReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	client := &http.Client{}
	listResp, err := client.Do(listReq)
	if err != nil {
		return fmt.Errorf("failed querying ranges: %v", err)
	}
	defer listResp.Body.Close() //nolint:errcheck

	listBody, err := io.ReadAll(listResp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response: %v", err)
	}
	if listResp.StatusCode != 200 {
		return fmt.Errorf("failed querying ranges status_code=%d, body=%s", listResp.StatusCode, string(listBody))
	}

	var results []map[string]interface{}
	if err := json.Unmarshal(listBody, &results); err != nil {
		return fmt.Errorf("unable to unmarshal response: %v", err)
	}
	if len(results) == 0 {
		return fmt.Errorf("no ip range found with name %q", name)
	}
	rangeID := int(results[0]["id"].(float64))

	// Fetch full range including stats.
	getURL := fmt.Sprintf("%s/api/v1/ranges/%d", cfg.Url, rangeID)
	getReq, err := http.NewRequest("GET", getURL, nil)
	if err != nil {
		return fmt.Errorf("failed creating request: %v", err)
	}
	getReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
	getResp, err := client.Do(getReq)
	if err != nil {
		return fmt.Errorf("failed fetching range stats: %v", err)
	}
	defer getResp.Body.Close() //nolint:errcheck

	getBody, err := io.ReadAll(getResp.Body)
	if err != nil {
		return fmt.Errorf("unable to read response: %v", err)
	}
	if getResp.StatusCode != 200 {
		return fmt.Errorf("failed fetching range status_code=%d, body=%s", getResp.StatusCode, string(getBody))
	}

	var rang map[string]interface{}
	if err := json.Unmarshal(getBody, &rang); err != nil {
		return fmt.Errorf("unable to unmarshal response: %v", err)
	}

	d.SetId(fmt.Sprintf("%d", rangeID))
	_ = d.Set("cidr", rang["cidr"].(string))

	if statsRaw, ok := rang["stats"]; ok && statsRaw != nil {
		stats := statsRaw.(map[string]interface{})
		_ = d.Set("total_addresses", int(stats["total_addresses"].(float64)))
		_ = d.Set("used_addresses", int(stats["used_addresses"].(float64)))
		_ = d.Set("free_addresses", int(stats["free_addresses"].(float64)))
		_ = d.Set("utilization_pct", stats["utilization_pct"].(float64))
	}

	return nil
}
