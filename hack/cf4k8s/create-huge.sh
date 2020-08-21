#!/usr/bin/env bash

set -euo pipefail

ROOT="$(cd $(dirname $0) && pwd)"

source "${ROOT}/methods.sh"

# ENV
CLUSTER_NAME=${CLUSTER_NAME:-$1}
CF_DOMAIN=${CF_DOMAIN:-$CLUSTER_NAME.k8s.capi.land}
: "${SHARED_DNS_ZONE_NAME:="kubenetes-clusters"}"
: "${GKE_GCP_PROJECT:="cf-capi-arya"}"
: "${DNS_GCP_PROJECT:="cff-capi-dns"}"


function main() {
  create_and_target_huge_cluster
  deploy_cf_for_k8s
  configure_dns
  target_cf
  enable_docker
}

main
