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

* This README takes you through the steps of running the DSS on Kubernetes and assumes familiarity with Kubernetes. It is also recommended to read up on CockroachDB, although that is not a requirement.

0. Clone/Fork this repo
1. Run `export NAMESPACE=YOUR_KUBERNETES_NAMESPACE`.
2. Make sure your `$KUBECONFIG` is pointed to the proper Kubernetes cluster and your context is set accordingly.
3. Ensure cockroach binary is installed and is able to be run with `cockroach version`. We use this to generate the certs in the python script.
4. Uncomment and fill out the **create_clusters** section in `make-certs.py` with the Kube namespace only.
  * If you are joining existing clusters, make sure to fill in the join_clusters variable with the appropriate node addresses and the path to their public cert.

   Run the `make-certs.py` script to generate the certs

   > `python2.7 make-certs.py`

   This script does 2 things:

- Builds a certificate directory structure
- Creates certificates within their respective directory to be used by the `apply-certs.sh` script.

5. Run `apply-certs.sh`. This script will delete existing secrets on the cluster named `cockroachdb.client.root` and `cockroachdb.node` and create secrets on the cluster containing the certificates that were generated from the python script.
6. Build the docker images for both the gRPC backend and HTTP Gateway. From the project's root directory:
  * $ `docker build -f cmds/grpc-backend/Dockerfile -t grpc-backend .`
  * $ `docker build -f cmds/http-gateway/Dockerfile -t http-gateway .`
  * Tag and push both containers to a Docker repo that your kubernetes cluster can access.
7. Fill out the `values.yaml` file with at minimum the ips, namespace, and storageClass, backendImage, and gatewayImage values.
8. Run `helm template . > cockroachdb.yaml` to render the YAML.
9. Run `kubectl apply -f cockroackdb.yaml` to apply it to the cluster.
10. Now that you have some pods, run `./expose.sh` to create an external IP for each pod.
11. Make sure that all of your necessary IP's are static.
12. Fill in these new IP addresses in make-certs and values.yaml file. Repeat steps 5-9.
  * TODO: automate this so we don't have to repeat these steps.
