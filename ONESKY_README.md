# OneSky Specific DSS Readme

## DSS Deployment Overview

This guide assumes a fully configured Kubernetes cluster that has the capability to deploy applications.

Required components in the cluster are nginx-ingress-controller, metrics server and cert-manager. This also requires a route53 DNS setup, example of dss.oneskysystems.com. Cluster setup walkthrough is provided here: https://github.com/OneSkySystems/kubernetes-deployment

## Deploy DSS instance to Kubernetes

- Make sure you have cockroachdb, tanka, kubectl, docker installed on your local machine. If the workspace folder already exists, you can skip some of the commands below. When in doubt, re-run them to overwrite the existing values AFTER taking a backup.

- Build the latest images using dss/build/build.sh (if intending to update the cluster)

    - Execute the following Jenkins job with the latest code pulled from the interuss/dss fork: https://build.onesky.xyz/job/EP_Migration/job/build_dss_image_main/

- Edit main.jsonnet and spec.json in the dss/build/workspace/$CLUSTER_CONTEXT folder. If the workspace folder already exists, the VAR_* values should be set. Any missing VAR_* values should be updated and set to their correct value.

    - Note, this shouldn't need to be updated for an update or redeployment, but please review the values to make sure they are correct.

    - The docker images may need to be updated, search for the ECR URL to find the image locations. The new images will be the result of the build in the previous step. This step is only necessary when updating the version of the DSS.

    - If the audience configuration is no longer "localhost" or needs to be updated, locate the `accepted_jwt_audiences` field and make the relevant updates

- Execute make-certs.py using the following command. Note, inside make-certs.py are the list of certificate SANS that are available. Depending on what the main.jsonnet file has configured for `hostnameSuffix`, you want to make sure the hostname is included in the list of SANS. This will generate the certs that will be deployed to kubernetes.

        python3 make-certs.py --cluster arn:aws:eks:us-east-2:169922227793:cluster/dss-ohio --namespace default

- Provide the CA and DSS certificates to cockroachdb using `apply-certs.sh`. This pushes the generated certificates into Kubernetes.

        ./apply-certs.sh arn:aws:eks:us-east-2:169922227793:cluster/dss-ohio default

- Restart cockroachDB pods using a rolling-restart to pick up new certificates. This can take up to 10 minutes.

        kubectl rollout restart statefulset/cockroachdb --namespace default

- Execute the following command to deploy the entire DSS deployment

        # Ex. ./build/workspacearn:aws:eks:us-east-2:169922227793:cluster
        tk apply .
        # If this is the initial installation, to get istio to install properly you'll need to run this twice. This seems to be a bug in the DSS deployment configuration.

    - If something goes wrong, you can delete using `tk delete .`

    - Note: Due to a mis-configuration with istio in the project, before deleting a DSS deployment, you MAY need to set enable_istio=false in main.jsonnet within the workspace directory when running `tk delete .` This will require deleting the istio-system namespace manually.

    - Note: Sometimes if something does not update from a configuration change, it will need to be destroyed and recreated to force the change.

- Execute the following kubernetes configuration files to set up networking on the cluster. These configuration files are onesky and cluster-specific, they may need to be updated to point to a different location.

        cd dss/onesky
        kubectl apply -f .

    - This will deploy ingresses and services that will make all of the components in the cluster accessible outside the EKS cluster. This is a 1:1 replace for the http-ingress GKE configuration included with the libsonnet deployment configuration. This means that without these yamls applied to the cluster, there is no direct access to the DSS deployment.

- The system is now up and minimally functioning.

## Istio Notes

As of the writing of this document, the istio-auto-sidecar-injector is not working correctly. Therefore, to enable istio benefits, the libsonnet files have been updated with a hard-coded `"sidecar.istio.io/inject": "true"`. To remove istio, perform a find and replace and change `true` to `false` to remove any clashing or issues. Here is the annotation configured in full, scattered through the libsonnet files.

        metadata+: {
            annotations+: {
              "sidecar.istio.io/inject": "true",
            },
          },

One of the more useful containers to annotate and track is the nginx-ingress-controller. However, that's managed by helm so we can't edit it's creation. We can however add it into a running cluster by adding an annotation at run-time using:

        kubectl annotate pods ingress-nginx-controller-<pod-id> sidecar.istio.io/inject=true -n ingress-nginx

Unfortunately this would need to be done each time the nginx-ingress-controller is updated or altered in any way, but it's not a required addition. Once this is completed, we'll need to scale or otherwise restart the ingress controller. This can be done with zero downtime by scaling up to N>1 pods, and destroying the original pod.

If the DSS components do not start up with the istio sidecar container, you must either scale or restart each pod individually. On recreation, it will deploy the sidecar successfully. This can be done by scaling the deployment, or killing the pods one at a time. NOTE, do not do this on the cockroachDB pods as they are dependent on quorum to continue functioning.

### Accessing the DSS

All URLs and data can be retrieved using the Kubernetes CLI. For example, to find the URL that is being exposed via ingress, run the following command (example output included):

        ➜ kubectl get ingress
        NAME                          CLASS    HOSTS
        onesky-grafana-ingress        <none>   dss-ohio.oneskysystems.com
        onesky-http-gateway-ingress   <none>   dss-ohio.oneskysystems.com

### Logging into Kiali

You must have working credentials into the target kubernetes cluster to access Kiali.

Because Kiali is using default credentials, it is only accessible via direct access through the kubernetes cluster. This can be enabled by executing a kubernetes port-forward.

        # Execute to get the pod name
        $ kubectl get pods -n istio-system
        NAME                                      READY   STATUS
        kiali-579dd86496-6w6xw                    1/1     Running
        $ kubectl port-forward kiali-579dd86496-6w6xw 20001:20001 -n istio-system
        Forwarding from 127.0.0.1:20001 -> 20001
        Forwarding from [::1]:20001 -> 20001

With this actively executing, you can navigate your browser to http://localhost:20001/kiali which will direct you to the pod inside the cluster.

This is not accessible outside of the cluster or a port-forward.

### Logging into Grafana

Grafana currently is set up with non-standard credentials, but can be updated to connect into Github or have dedicated accounts (among other auth/authn methods).

Grafana is available at https://dss-ohio.oneskysystems.com/grafana/ with onesky/\<password-available on request>.

### Debugging Cockroach DB

Exec into one of the cockroach containers using `kubectl exec -it cockroachdb-1 bash`.

From there you can use the ./cockroach command based on the documentation to query cockroachdb.

You can also port-forward into a cockroachdb container and review the UI at http://localhost:8080 after executing `kubectl port-forward cockroachdb-1 8080:8080`. Having a created non-root user is required to get UI access, so we can execute the following to create this basic user.

        kubectl exec -it cockroachdb-1 bash
        ./cockroach --certs-dir=/cockroach/cockroach-certs/ sql
        CREATE USER "test" WITH PASSWORD "Dr0ne$";

#### Cockroach notable issues

If one of the cockroachDBs has issues, sometimes you need to delete the pod AS WELL as the PVC to fully destroy the existing CRDB instance and configuration.

#### Recover a dead CockroachDB Instance

If one of the nodes dies, the DSS configuration is not sufficient for easy recovery. These operations below will recover the cluster.

Please review the disaster-recovery page for cockroachdb for background on what might cause an instance to become "DEAD": https://www.cockroachlabs.com/docs/v21.1/disaster-recovery.html

0. This method is spotty and unreliable, disaster recovery should be investigated further.

1. Review the cluster status: `./cockroach node status --certs-dir=/cockroach/cockroach-certs/`

2. Locate the damaged node ID and execute the following command `./cockroach node decommission <damaged-node-id-integer> --certs-dir=/cockroach/cockroach-certs/`

3. Once this is finished executing (you can verify this through the CockroachDB UI mentioned above), you can re-execute `tk apply .` in the workspace/\<context> directory to trigger an reconfigure of cockroachDB.
