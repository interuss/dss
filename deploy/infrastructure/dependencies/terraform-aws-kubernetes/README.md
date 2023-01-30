# terraform-aws-kubernetes

This module creates an EKS cluster in AWS. 

## Getting started

1. Install prerequisites:
   1. `aws cli`: https://docs.aws.amazon.com/cli/latest/userguide/getting-started-install.html#getting-started-install-instructions
2. If you don't have an account, sign-up: https://aws.amazon.com/free/
3. Configure terraform to connect to AWS using your account. 
   1. We recommend to create an AWS_PROFILE using for instance `aws configure --profile aws-interuss-dss`
   Before running `terraform` commands, run once in your shell: `export AWS_PROFILE=aws-interuss-dss`
   Other methods are described here: https://registry.terraform.io/providers/hashicorp/aws/latest/docs#authentication-and-configuration
4. Create a SSH keypair to connect to your worker nodes using SSH. 
   1. Connect to the [AWS console](https://console.aws.amazon.com).
   2. Go to the EC2 page.
   3. Click on the Key Pairs section.
   4. Make sure you are working in the correct region. 
   5. Add or import your SSH key: https://console.aws.amazon.com/ec2/home#KeyPairs: 

## Provision the infrastructure

1. Copy and .....
terraform apply 
Creation takes 20 min.

Login = aws eks --region eu-west-1 update-kubeconfig --name dss-1


## Design

This module creates an EKS cluster with related worker nodes. 
EKS requires 2 subnets in different availability zones (AZ) is provisioned. 
A dedicated VPC is created to that effect.
At the moment, worker nodes are deployed in the two first AZ of the region.

The following table summarizes current responsibilities for resource creation in the AWS account:

| Resource type                               | Manager                      |
|---------------------------------------------|------------------------------|
| VPC and Subnets                             | Terraform                    |
| Elastic IPs                                 | Terraform                    |
| Network Load balancer                       | aws-load-balancer-controller |
| Target groups                               | aws-load-balancer-controller |
| SSL Certificates (AWS Certificates Manager) | Terraform                    |
| DNS                                         | Terraform (or manual)        |


### AWS Load Balancers and Kubernetes Services

Load balancers are provisioned by the aws-load-balancer-controller v2.4 with [Option B for IAM configuration](https://kubernetes-sigs.github.io/aws-load-balancer-controller/v2.4/deploy/installation/#option-b-attach-iam-policies-to-nodes).

Network Load Balancers map elastic IPs to Kubernetes Service (Load Balancer). 
Application Load Balancers (Ingress) do not support this feature at the moment, making impossible to anticipate DNS records.

The Network Load Balancers are provisioned by the aws-load-balancer-controller.
It can handle the TLS termination. (For the dss https service for instance)

Note that the load balancer is distributing the traffic to possibly multiple subnets. 
Be aware that it is not possible to unassign a subnet. Target pods shall be deployed in
every subnet, meaning that the pods should be properly distributed in worker nodes and 
a worker node should be at least present in each subnets.

Provisioning is done by annotating a Kubernetes Service resource. The following example deploys a simple http server:
```yaml
---
apiVersion: v1
kind: Namespace
metadata:
  name: example
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example-app
  namespace: example
  labels:
    app: example-app
spec:
  replicas: 2
  selector:
    matchLabels:
      app: example-app
  template:
    metadata:
      labels:
        app: example-app
    spec:
      containers:
        - name: nginx
          image: public.ecr.aws/nginx/nginx:1.21
          ports:
            - name: http
              containerPort: 80
          imagePullPolicy: IfNotPresent
---
apiVersion: v1
kind: Service
metadata:
  name: example-service
  namespace: example
  annotations:
    service.beta.kubernetes.io/aws-load-balancer-type: external
    service.beta.kubernetes.io/aws-load-balancer-nlb-target-type: ip
    service.beta.kubernetes.io/aws-load-balancer-scheme: "internet-facing"
    service.beta.kubernetes.io/aws-load-balancer-ssl-ports: '443'
    service.beta.kubernetes.io/aws-load-balancer-ssl-cert: [CERTIFICATE_ARN]
    service.beta.kubernetes.io/aws-load-balancer-eip-allocations: [EIP_ALLOCATION_ID1,EIP_ALLOCATION_ID2,...]
    service.beta.kubernetes.io/aws-load-balancer-name: [LOAD_BALANCER_NAME]
    service.beta.kubernetes.io/aws-load-balancer-subnets: [SUBNET_ID1,SUBNET_ID2,...]
spec:
  selector:
    app: example-app
  ports:
    - port: 443
      targetPort: 80
      protocol: TCP
      name: http
  type: LoadBalancer
  loadBalancerClass: service.k8s.aws/nlb
```

- [CERTIFICATE_ARN]: arn of the certificate managed by AWS Certificate Manager
- [EIP_ALLOCATION_IDx]: Elastic IP allocation id (The number of elastic IP should equal to the number of SUBNET_IDx)
- [LOAD_BALANCER_NAME]: Name of the balancer created by the controller
- [SUBNET_IDx]: Name or ID of a subnet (The number of subnets should equal to the number of EIP_ALLOCATION_IDx)

## Test
`terraform apply` generates a resource specification `test-app.yml`. This file can be applied to test a http server
reachability within the deployed cluster. To apply the resources, follow the next steps:

1. Login to the EKS cluster: `aws eks --region $AWS_REGION update-kubeconfig --name $CLUSTER_NAME`
2. Create the resources: `kubectl apply -f test-app.yml`
3. Wait (up to 5min) for the load balancer to be ready. Note that the load balancer may take few minutes to start. 
Monitor the progress here until the state becomes `Active`: https://console.aws.amazon.com/ec2/home#LoadBalancers:
4. Connect to the app_hostname and contemplate the nginx default welcome page.

## Clean up

Delete all services in kubernetes.
Make sure all load balancers and target groups have been removed.
terraform destroy

