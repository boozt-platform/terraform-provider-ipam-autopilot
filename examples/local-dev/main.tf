# Copyright 2026 Boozt Fashion AB
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Local development example — requires docker compose up to be running.
#
# Setup:
#   1. docker compose up --build -d
#   2. cd provider && go build -o terraform-provider-ipam-autopilot .
#   3. Add dev_overrides to ~/.terraformrc (see below)
#   4. export GCP_IDENTITY_TOKEN=localdev
#   5. terraform plan / apply (no init needed with dev_overrides)
#
# ~/.terraformrc:
#   provider_installation {
#     dev_overrides {
#       "boozt-platform/ipam-autopilot" = "/path/to/repo/provider"
#     }
#     direct {}
#   }

terraform {
  required_providers {
    ipam = {
      source = "boozt-platform/ipam-autopilot"
    }
  }
}

provider "ipam" {
  url = "http://localhost:8080"
}

resource "ipam_routing_domain" "local" {
  name = "local-dev"
  vpcs = []
}

resource "ipam_ip_range" "parent" {
  name       = "local-parent"
  range_size = 8
  cidr       = "10.0.0.0/8"
  domain     = ipam_routing_domain.local.id
}

resource "ipam_ip_range" "gke_nodes" {
  name       = "local-gke-nodes"
  range_size = 22
  parent     = ipam_ip_range.parent.cidr
  domain     = ipam_routing_domain.local.id
  labels     = { env = "local", purpose = "gke-nodes" }
}

resource "ipam_ip_range" "gke_pods" {
  name       = "local-gke-pods"
  range_size = 16
  parent     = ipam_ip_range.parent.cidr
  domain     = ipam_routing_domain.local.id
  labels     = { env = "local", purpose = "gke-pods" }
}

resource "ipam_ip_range" "gke_services" {
  name       = "local-gke-services"
  range_size = 20
  parent     = ipam_ip_range.parent.cidr
  domain     = ipam_routing_domain.local.id
  labels     = { env = "local", purpose = "gke-services" }
}

resource "ipam_ip_range" "mgmt" {
  name       = "local-mgmt"
  range_size = 24
  parent     = ipam_ip_range.parent.cidr
  domain     = ipam_routing_domain.local.id
  labels     = { env = "local", purpose = "mgmt" }
}

data "ipam_ip_range" "gke_nodes_lookup" {
  id = ipam_ip_range.gke_nodes.id
}

data "ipam_ip_range_stats" "parent_stats" {
  id = ipam_ip_range.parent.id
  depends_on = [
    ipam_ip_range.gke_nodes,
    ipam_ip_range.gke_pods,
    ipam_ip_range.gke_services,
    ipam_ip_range.mgmt,
  ]
}

output "allocated_cidrs" {
  value = {
    gke_nodes    = ipam_ip_range.gke_nodes.cidr
    gke_pods     = ipam_ip_range.gke_pods.cidr
    gke_services = ipam_ip_range.gke_services.cidr
    mgmt         = ipam_ip_range.mgmt.cidr
  }
}

output "gke_nodes_cidr_via_datasource" {
  value = data.ipam_ip_range.gke_nodes_lookup.cidr
}

output "parent_stats" {
  value = {
    total_addresses = data.ipam_ip_range_stats.parent_stats.total_addresses
    used_addresses  = data.ipam_ip_range_stats.parent_stats.used_addresses
    free_addresses  = data.ipam_ip_range_stats.parent_stats.free_addresses
    utilization_pct = data.ipam_ip_range_stats.parent_stats.utilization_pct
  }
}
