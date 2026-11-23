# Decomissioning a DSS instance deployed with terraform

## Clean up

!!! danger
    Note that the following operations can't be reverted and all data will be lost.

1. Navigate to the workspace folder and run `./get-credentials.sh` to login to kubernetes; see corresponding [services deployment documentation](../deployment/services/after-terraform.md)
2. To delete all resources, run `tk delete .` in the tanka workspace folder (Tanka) or `helm uninstall --kube-context=${cluster_context} dss` (Helm).
3. (AWS) Make sure that all [load balancers](https://eu-west-1.console.aws.amazon.com/ec2/home#LoadBalancers:) and [target groups](https://eu-west-1.console.aws.amazon.com/ec2/home#TargetGroups:) have been deleted from the AWS region before next step.
4. (GCP) If `google_delete_protection` is `true` (default), set it to `false` in your `terraform.tfvars` and run `terraform apply` in your infrastructure folder, otherwise `terraform destroy` fails. This protection is stored in the Terraform state only and can't be changed from the Google Cloud console.
5. `terraform destroy` in your infrastructure folder.
6. (AWS) On the [EBS page](https://eu-west-1.console.aws.amazon.com/ec2/home#Volumes:), make sure to manually clean up the persistent storage. Note that the correct AWS region shall be selected.
7. (GCP) [Manually clean up the persistent storage](https://console.cloud.google.com/compute/disks)
