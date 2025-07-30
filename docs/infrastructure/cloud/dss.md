# Deploy DSS services

During the successful run, the terraform job has created a new [workspace](../../../../build/workspace/)
for the cluster. The new workspace name corresponds to the cluster context. The cluster context
can be retrieved by running `terraform output` in your infrastructure  folder (ie /deploy/infrastructure/personal/terraform-aws-dss-dev).

It contains scripts to operate the cluster and setup the services.

1. Go to the new workspace `/build/workspace/${cluster_context}`.
2. Run `./get-credentials.sh` to login to kubernetes. You can now access the cluster with `kubectl`.
3. If using CockroachDB:
    1. Generate the certificates using `./make-certs.sh`. Follow script instructions if you are not initializing the cluster.
    1. Deploy the certificates using `./apply-certs.sh`.
4. If using Yugabyte:
    1. Generate the certificates using `./dss-certs.sh init`
    1. If joining a cluster, check `dss-certs.sh`'s [help](../../../operations/certificates-management/README.md) to add others CA in your pool and share your CA with others pools members.
    1. Deploy the certificates using `./dss-certs.sh apply`.
5. Run `tk apply .` to deploy the services to kubernetes. (This may take up to 30 min)
6. Wait for services to initialize:
    - On AWS, load balancers and certificates are created by Kubernetes Operators. Therefore, it may take few minutes (~5min) to get the services up and running and generate the certificate. To track this progress, go to the following pages and check that:
        - On the [EKS page](https://eu-west-1.console.aws.amazon.com/eks/home), the status of the kubernetes cluster should be `Active`.
        - On the [EC2 page](https://eu-west-1.console.aws.amazon.com/ec2/home#LoadBalancers:), the load balancers (1 for the gateway, 1 per cockroach nodes) are in the state `Active`.
    - On Google Cloud, the highest-latency operation is provisioning of the HTTPS certificate which generally takes 10-45 minutes. To track this progress:
        - Go to the "Services & Ingress" left-side tab from the Kubernetes Engine page.
        - Click on the https-ingress item (filter by just the cluster of interest if you have multiple clusters in your project).
        - Under the "Ingress" section for Details, click on the link corresponding with "Load balancer".
        - Under Frontend for Details, the Certificate column for HTTPS protocol will have an icon next to it which will change to a green checkmark when provisioning is complete.
        - Click on the certificate link to see provisioning progress.
        - If everything indicates OK and you still receive a cipher mismatch error message when attempting to visit /healthy, wait an additional 5 minutes before attempting to troubleshoot further.
7. Verify that basic services are functioning by navigating to https://your-gateway-domain.com/healthy.

