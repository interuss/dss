# Setup DNS

This page describes the options and steps required to setup DNS for a DSS deployment.

## Terraform managed

If your DNS zone is managed on the same account, it is possible to instruct terraform to create and manage
it with the rest of the infrastructure.

=== "AWS"
    **For Elastic Kubernetes Service (AWS)**, create the zone in your aws account and set the `aws_route53_zone_id`
    variable with the zone id. Entries will be automatically created by terraform.
     Note that the domain or the sub-domain managed by the zone must be properly delegated by the parent domain.
    See instructions for [subdomains delegation](https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/CreatingNewSubdomain.html#UpdateDNSParentDomain)

=== "GCP"
    **For Google Cloud Engine**, configure the zone in your google account and set the `google_dns_managed_zone_name`
    variable the zone name. Zones can be listed by running `gcloud dns managed-zones list`. Entries will be
    automatically created by terraform.


## Manual setup

If DNS entries are managed manually, set them up manually using the following steps:

### AWS
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
       "certificate_validation_dns" = [
        {
          "managed_by_terraform" = false
          "name" = "_6e246283563dcf58e7ed.interuss.example.com."
          "records" = [
             "_6e246283563dcf58e7ed.xxxxx.acm-validations.aws.",
          ]
          "type" = "CNAME"
        },
       ]
   }
   ```
2. Create the following DNS A entries to point to the static ips:
    - `crdb_addresses[*].expected_dns`
    - `gateway_address.expected_dns`

3. Create the entries for SSL certificate validation according to the information provided
    in `gateway_address.certificate_validation_dns`.


### GCP
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
    ```
2. Create the related DNS A entries to point to the static ips.
