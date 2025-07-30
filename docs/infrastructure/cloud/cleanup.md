# Clean up
Note that the following operations can't be reverted and all data will be lost.

## AWS
3. To delete all resources, run `tk delete .` in the workspace folder.
3. Make sure that all [load balancers](https://eu-west-1.console.aws.amazon.com/ec2/home#LoadBalancers:) and [target groups](https://eu-west-1.console.aws.amazon.com/ec2/home#TargetGroups:) have been deleted from the AWS region before next step.
4. `terraform destroy` in your infrastructure folder.
5. On the [EBS page](https://eu-west-1.console.aws.amazon.com/ec2/home#Volumes:), make sure to manually clean up the persistent storage. Note that the correct AWS region shall be selected.

## GCP
To delete all resources, run `terraform destroy`. Note that this operation can't be reverted and all data will be lost.

For Google Cloud Engine, make sure to manually clean up the persistent storage: https://console.cloud.google.com/compute/disks
