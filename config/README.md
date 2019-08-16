# Multi-region cockroachdb setup

## Prerequisites:
* Download & install:
 - helm
 - kubectl
 - docker
 - cockroachdb
 - [Optional] Golang. Recommended to understand go, and the go toolchain.
 - [Optional] Run `make all` in the top level directory
 - [Optional] minikube, docker-for-mac, or other local kubernetes system.

0. Run `export NAMESPACE=YOUR_KUBERNETES_NAMESPACE`.
1. Make sure your `$KUBECONFIG` is pointed to the proper Kubernetes cluster and your context is set accordingly.
2. Ensure cockroach binary is installed and is able to be run with `cockroach version`. We use this to generate the certs in the python script.
3. Install helm because it is needed to generate the templates. (would like to transition to kustomize later on since its natively supported in kubectl)
4. Be sure to run all scripts from within this directory.
5. Uncomment and fill out the **create_clusters** section in the python script with a namespace and context.
  * If you are joining existing clusters, make sure to fill in the join_clusters variable with the appropriate node addresses and the path to their public cert.

   Run the `make-certs.py` script to generate the certs

   > `python2.7 make-certs.py`

   This script does 2 things:
   This script does 3 things:

- Builds a certificate directory structure
- Creates certificates within their respective directory to be used by the `apply-certs.sh` script.

6. Fill out the `NAMESPACE` and `CLUSTER_INIT` variables at the top of `apply-certs.sh` and then run it to load the secrets into the script. This script will delete existing secrets on the cluster named `cockroachdb.client.root` and it will also create secrets on the cluster containing the certificates that were generated from the python script.
7. Fill out the `values.yaml` file with at minimum the ips, namespace, and storageClass values.
8. Run `helm template . > cockroachdb.yaml` to render the YAML.
9. Run `kubectl apply -f cockroackdb.yaml` to apply it to the cluster.
10. Now that you have some pods, run `./expose.sh` to create an external IP for each pod.
11. Make sure that all of your necessary IP's are static.
12. Fill in these new IP addresses in make-certs and values.yaml file. Repeat steps 5-9.
  * TODO: automate this so we don't have to repeat these steps.
