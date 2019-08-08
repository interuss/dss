# Multi-region cockroachdb setup

1. Make sure your `$KUBECONFIG` is pointed to the proper cluster and your context is set accordingly.
2. Ensure cockroach binary is installed and is able to be run with `cockroach version`. We use this to generate the certs in the python script.
3. Install helm because it is needed to generate the templates. (would like to transition to kustomize later on since its natively supported in kubectl)
4. Be sure to run all scripts from within this directory.
5. Uncomment and fill out the **create_clusters** section in the python script with a namespace and context.

   Run the `make-certs.py` script to generate the certs

   > `python2.7 make-certs.py`

   This script does 3 things:

- Generates a public-facing loadbalancer in the designated namespace.
- Builds a certificate directory structure
- Creates certificates within their respective directory to be used by the `apply-certs.sh` script.

5. Fill out the `NAMESPACE` and `CLUSTER_INIT` variables at the top of `apply-certs.sh` and then run it to load the secrets into the script. This script will delete existing secrets on the cluster named `cockroachdb.client.root` and it will also create secrets on the cluster containing the certificates that were generated from the python script.
6. Fill out the `values.yaml` file with at minimum the PublicAddr, namespace, and storageClass values.
7. Run `helm template . > cockroachdb.yaml` to render the YAML.
8. Run `kubectl apply -f cockroackdb.yaml` to apply it to the cluster.
