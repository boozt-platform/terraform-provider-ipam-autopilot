// Copyright 2021 Google LLC
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

package main

import (
	"context"
	"fmt"
	"log"
	"reflect"

	asset "cloud.google.com/go/asset/apiv1"
	assetpb "cloud.google.com/go/asset/apiv1/assetpb"
	"google.golang.org/api/iterator"
)

type CaiSecondaryRange struct {
	name string
	cidr string
}
type CaiRange struct {
	name            string
	id              string
	network         string
	cidr            string
	secondaryRanges []CaiSecondaryRange
}

func GetRangesForNetwork(parent string, networks []string) ([]CaiRange, error) {
	ctx := context.Background()
	client, err := asset.NewClient(ctx)
	if err != nil {
		return nil, err
	}

	defer client.Close() //nolint:errcheck

	itr := client.ListAssets(ctx, &assetpb.ListAssetsRequest{
		Parent:      parent,
		AssetTypes:  []string{"compute.googleapis.com/Subnetwork"},
		ContentType: assetpb.ContentType_RESOURCE,
	})

	ranges := make([]CaiRange, 0)

	for asset, err := itr.Next(); err != iterator.Done; asset, err = itr.Next() {
		if err != nil {
			log.Fatal(err)
		}
		if containsValue(networks, asset.Resource.Data.Fields["network"].GetStringValue()) {
			secondaryRanges := make([]CaiSecondaryRange, 0)
			secondary := asset.Resource.Data.Fields["secondaryIpRanges"].GetListValue().AsSlice()
			for i := 0; i < len(secondary); i++ {
				var rangeName string
				var ipCidrRange string

				iter := reflect.ValueOf(secondary[i]).MapRange()
				for iter.Next() {
					key := iter.Key().Interface()
					value := iter.Value().Interface()
					if key == "ipCidrRange" {
						ipCidrRange = fmt.Sprintf("%s", value)
					}
					if key == "rangeName" {
						rangeName = fmt.Sprintf("%s", value)
					}
				}
				secondaryRanges = append(secondaryRanges, CaiSecondaryRange{
					name: rangeName,
					cidr: ipCidrRange,
				})
			}
			ranges = append(ranges, CaiRange{
				id:              asset.Resource.Data.Fields["id"].GetStringValue(),
				name:            asset.Name,
				network:         asset.Resource.Data.Fields["network"].GetStringValue(),
				cidr:            asset.Resource.Data.Fields["ipCidrRange"].GetStringValue(),
				secondaryRanges: secondaryRanges,
			})
		} else {
			log.Printf("Ignoring network %s", asset.Resource.Data.Fields["network"].GetStringValue())
		}
	}
	return ranges, nil
}

func containsValue(array []string, lookup string) bool {
	for i := 0; i < len(array); i++ {
		if lookup == array[i] {
			return true
		}
	}
	return false
}
