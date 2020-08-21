#!/usr/bin/env bash


#!/usr/bin/env bash

set -euo pipefail

# ENV
CLUSTER_NAME=${CLUSTER_NAME:-$1}
CF_DOMAIN=${CF_DOMAIN:-$CLUSTER_NAME.k8s.capi.land}
: "${SHARED_DNS_ZONE_NAME:="kubenetes-clusters"}"
: "${GKE_GCP_PROJECT:="cf-capi-arya"}"
: "${DNS_GCP_PROJECT:="cff-capi-dns"}"

function delete_cluster() {
    if gcloud container clusters describe --project ${GKE_GCP_PROJECT} --zone us-west1-a ${CLUSTER_NAME} > /dev/null; then
        echo "Deleting cluster: ${CLUSTER_NAME} ..."
        gcloud container clusters delete ${CLUSTER_NAME} --project ${GKE_GCP_PROJECT} --zone us-west1-a
    else
        echo "${CLUSTER_NAME} already deleted! Continuing..."
    fi
}

function delete_dns() {
  echo "Deleting DNS for: *.${CF_DOMAIN}"
  gcloud dns record-sets transaction start --project ${DNS_GCP_PROJECT} --zone="${SHARED_DNS_ZONE_NAME}"
  gcp_records_json="$( gcloud dns record-sets list --project ${DNS_GCP_PROJECT} --zone "${SHARED_DNS_ZONE_NAME}" --name "*.${CF_DOMAIN}" --format=json )"
  record_count="$( echo "${gcp_records_json}" | jq 'length' )"
  if [ "${record_count}" != "0" ]; then
    existing_record_ip="$( echo "${gcp_records_json}" | jq -r '.[0].rrdatas | join(" ")' )"
    gcloud dns record-sets transaction remove --name "*.${CF_DOMAIN}" --type=A --project ${DNS_GCP_PROJECT} --zone="${SHARED_DNS_ZONE_NAME}" --ttl=300 "${existing_record_ip}" --verbosity=debug
  fi

  echo "Contents of transaction.yaml:"
  cat transaction.yaml
  gcloud dns record-sets transaction execute --project ${DNS_GCP_PROJECT} --zone="${SHARED_DNS_ZONE_NAME}" --verbosity=debug
}

function cleanup() {
  rm -rf /tmp/${CF_DOMAIN}*
}

function main() {
    delete_dns
    delete_cluster
    cleanup
}

main
