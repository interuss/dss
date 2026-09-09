# Decomissioning a DSS instance deployed with terraform

## Clean up

!!! danger
    Note that the following operations can't be reverted and all data will be lost.

1. Navigate to the workspace folder and run `./get-credentials.sh` to login to kubernetes; see corresponding [services deployment documentation](../deployment/services/after-terraform.md)
2. To delete all resources, run `tk delete .` in the workspace folder.
3. (AWS) Make sure that all [load balancers](https://eu-west-1.console.aws.amazon.com/ec2/home#LoadBalancers:) and [target groups](https://eu-west-1.console.aws.amazon.com/ec2/home#TargetGroups:) have been deleted from the AWS region before next step.
4. `terraform destroy` in your infrastructure folder.
5. (AWS) On the [EBS page](https://eu-west-1.console.aws.amazon.com/ec2/home#Volumes:), make sure to manually clean up the persistent storage. Note that the correct AWS region shall be selected.
6. (GCP) [Manually clean up the persistent storage](https://console.cloud.google.com/compute/disks)
