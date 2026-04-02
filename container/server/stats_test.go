// Copyright 2026 Boozt Fashion AB
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0

package server

import (
	"testing"
)

func TestComputeStats(t *testing.T) {
	tests := []struct {
		name      string
		cidr      string
		children  []string
		wantTotal int64
		wantUsed  int64
		wantFree  int64
		wantPct   float64
		wantErr   bool
	}{
		{
			name:      "no children",
			cidr:      "10.0.0.0/8",
			children:  nil,
			wantTotal: 16777216,
			wantUsed:  0,
			wantFree:  16777216,
			wantPct:   0,
		},
		{
			name:      "single /22 child in /16",
			cidr:      "10.0.0.0/16",
			children:  []string{"10.0.0.0/22"},
			wantTotal: 65536,
			wantUsed:  1024,
			wantFree:  64512,
			wantPct:   1.56,
		},
		{
			name:      "two /22 children in /8",
			cidr:      "10.0.0.0/8",
			children:  []string{"10.0.0.0/22", "10.0.4.0/22"},
			wantTotal: 16777216,
			wantUsed:  2048,
			wantFree:  16775168,
			wantPct:   0.01,
		},
		{
			name:      "fully allocated /24 with two /25s",
			cidr:      "10.0.0.0/24",
			children:  []string{"10.0.0.0/25", "10.0.0.128/25"},
			wantTotal: 256,
			wantUsed:  256,
			wantFree:  0,
			wantPct:   100,
		},
		{
			name:    "invalid cidr returns error",
			cidr:    "not-a-cidr",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stats, err := computeStats(tc.cidr, tc.children)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if stats.TotalAddresses != tc.wantTotal {
				t.Errorf("TotalAddresses: got %d, want %d", stats.TotalAddresses, tc.wantTotal)
			}
			if stats.UsedAddresses != tc.wantUsed {
				t.Errorf("UsedAddresses: got %d, want %d", stats.UsedAddresses, tc.wantUsed)
			}
			if stats.FreeAddresses != tc.wantFree {
				t.Errorf("FreeAddresses: got %d, want %d", stats.FreeAddresses, tc.wantFree)
			}
			if stats.UtilizationPct != tc.wantPct {
				t.Errorf("UtilizationPct: got %v, want %v", stats.UtilizationPct, tc.wantPct)
			}
		})
	}
}
