# Setup DNS

This page describes the options and steps required to setup DNS for a DSS deployment.

## Terraform managed

If your DNS zone is managed on the same account, it is possible to instruct terraform to create and manage
it with the rest of the infrastructure.

- **For Google Cloud Engine**, configure the zone in your google account and set the `google_dns_managed_zone_name`
  variable the zone name. Zones can be listed by running `gcloud dns managed-zones list`. Entries will be
  automatically created by terraform.

## Manual setup 

If DNS entries are managed manually, set them up manually using the following steps:

1. Retrieve IP addresses and expected hostnames: `terraform output`
   Example of expected output:
   ```
   crdb_addresses = [
       {
           "address" = "34.65.15.23"
           "expected_dns" = "0.interuss.example.com"
       },
       {
           "address" = "34.65.146.56"
           "expected_dns" = "1.interuss.example.com"
       },
       {
           "address" = "34.65.191.145"
           "expected_dns" = "2.interuss.example.com"
       },
   ]
   gateway_address = {
       "address" = "35.186.236.146"
       "expected_dns" = "dss.interuss.example.com"
   }
   
2. Create the related DNS A entries to point to the static ips.
