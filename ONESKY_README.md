# OneSky Specific DSS Readme

## Deploying a DSS to Kubernetes

This guide assumes a fully configured Kubernetes cluster that has the capability to deploy applications.

Required components in the cluster are nginx-ingress-controller, metrics server and cert-manager. This also requires a route53 DNS setup, example of dss.oneskysystems.com. Cluster setup walkthrough is provided here: https://github.com/OneSkySystems/kubernetes-deployment

### Deploy DSS instance to Kubernetes

###### Personal Note: cockroachdb-1.cockroachdb.default.svc.cluster.local not found, create

- Make sure you have cockroachdb, tanka, kubectl, docker installed on your local machine. If the workspace folder already exists, you can skip some of the commands below. When in doubt, re-run them to overwrite the existing values AFTER taking a backup.

- Build the latest images using dss/build/build.sh (if intending to update the cluster)

    - Execute the following Jenkins job with the latest code pulled from the interuss/dss fork: https://build.onesky.xyz/job/EP_Migration/job/build_dss_image_main/

- Edit main.jsonnet in the dss/build/workspace/$CLUSTER_CONTEXT folder. If the workspace folder already exists, the VAR_* values should be set. Any missing VAR_* values should be updated and set to their correct value.

    - Note, this shouldn't need to be updated for an update or redeployment, but please review the values to make sure they are correct.

- Execute make-certs.py using the following command. Note, inside make-certs.py are the list of certificate SANS that are available. Depending on what the main.jsonnet file has configured for `hostnameSuffix`, you want to make sure the hostname is included in the list of SANS. This will generate the certs that will be deployed to kubernetes.

        python3 make-certs.py --cluster arn:aws:eks:us-east-2:169922227793:cluster/dss-ohio --namespace default

- Provide the CA and DSS certificates to cockroachdb using `apply-certs.sh`. This pushes the generated certificates into Kubernetes.

        ./apply-certs.sh arn:aws:eks:us-east-2:169922227793:cluster/dss-ohio default

- Restart cockroachDB pods using a rolling-restart to pick up new certificates. This can take up to 10 minutes.

        kubectl rollout restart statefulset/cockroachdb --namespace default

- Execute the following command to deploy the entire DSS deployment

        # Ex. ./build/workspacearn:aws:eks:us-east-2:169922227793:cluster
        tk apply .

    - If something goes wrong, you can delete using `tk delete .`

- Execute the following kubernetes configuration files to set up networking on the cluster. These configuration files are onesky and cluster-specific, they may need to be updated to point to a different location.

        cd dss/onesky
        kubectl apply -f .

    - This will deploy ingresses and services that will make all of the components in the cluster accessible outside the EKS cluster. This is a 1:1 replace for the http-ingress GKE configuration included with the libsonnet deployment configuration. This means that without these yamls applied to the cluster, there is no direct access to the DSS deployment.


