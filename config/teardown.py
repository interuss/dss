#!/usr/bin/env python

from shutil import rmtree
from subprocess import call

import os

# Before running the script, fill in appropriate values for all the parameters
# above the dashed line. You should use the same values when tearing down a
# cluster that you used when setting it up.

# To get the names of your kubectl "contexts" for each of your clusters, run:
#   kubectl config get-contexts
contexts = {
    # 'us-east1-b': 'gke_wing-crdb_us-east1-b_cockroachdb1',
    # 'us-east4-a': 'gke_wing-crdb-2_us-east4-a_cockroachdb2',
    'us-central1-a': 'gke_wing-crdb-3_us-central1-a_cockroachdb3',
}

certs_dir = './certs'
ca_key_dir = './my-safe-directory'
generated_files_dir_tmpl = './generated'
# ------------------------------------------------------------------------------

# Delete each cluster's special zone-scoped namespace, which transitively
# deletes all resources that were created in the namespace, along with the few
# other resources we created that weren't in that namespace
for zone, context in contexts.items():
    generated_files_dir = os.path.join(generated_files_dir_tmpl, zone)
    call(['kubectl', 'delete', 'secret', 'cockroachdb.client.root', '--context', context])
    call(['kubectl', 'delete', 'secret', 'cockroachdb.client.root', '--namespace', zone, '--context', context])
    call(['kubectl', 'delete', 'secret', 'cockroachdb.node', '--namespace', zone,  '--context', context])
    call(['kubectl', 'delete', '-f', '%s/external-name-svc.yaml' % (generated_files_dir), '--context', context])
    call(['kubectl', 'delete', '-f', 'templates/dns-lb.yaml', '--context', context])
    call(['kubectl', 'delete', '-f', 'templates/cluster-init-secure.yaml', '--namespace', zone, '--context', context])
    call(['kubectl', 'delete', '-f', '%s/cockroachdb-statefulset-secure.yaml' % (generated_files_dir), '--namespace', zone, '--context', context])
    call(['kubectl', 'delete', 'configmap', 'kube-dns', '--namespace', 'kube-system', '--context', context])
    # Restart the DNS pods to clear out our stub-domains configuration.
    call(['kubectl', 'delete', 'pods', '-l', 'k8s-app=kube-dns', '--namespace', 'kube-system', '--context', context])
    call(['kubectl', 'delete', 'namespace', zone, '--context', context])