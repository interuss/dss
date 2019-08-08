#!/usr/bin/env python

import distutils.spawn
import json
import os
import stat
from subprocess import check_call, check_output
from sys import exit
from time import sleep

# This script builds a loadbalancer in the specified namespace and context and then generates certificates.


class CockroachCluster():

    def __init__(self, namespace, context='', lb_ip='', ca_certs_file=''):
        self.namespace = namespace
        self.context = context
        self.lb_ip = lb_ip
        if ca_certs_file:
            self.ca_certs_file = ca_certs_file

    @property
    def directory(self):
        return os.path.join('./generated/', self.namespace)

    @property
    def ca_certs_file(self):
        return os.path.join(self.ca_certs_dir, 'ca.crt')

    @property
    def ca_certs_dir(self):
        return os.path.join(self.directory, 'ca_certs_dir')

    @property
    def client_certs_dir(self):
        return os.path.join(self.directory, 'client_certs_dir')

    @property
    def node_certs_dir(self):
        return os.path.join(self.directory, 'node_certs_dir')

# EDIT/UNCOMMENT THIS!!

# create_clusters = [
#     CockroachCluster(
#         namespace='',
#         context='',
#     ),
# ]


join_clusters = [
    # CockroachCluster(
    #   lb_ip='external_ip_address',
    #   ca_certs_file='path_to_ca_public_cert',
    # ),
]

# Create cert folders, create the namespace, and apply the loadbalancer yaml.
for cr in create_clusters:
    try:
        os.mkdir('./generated')
    except OSError:
        pass
    try:
        os.mkdir(cr.directory)
    except OSError:
        pass
    try:
        check_call(['kubectl', 'create', 'namespace',
                    cr.namespace, '--context', cr.context])
    except:
        pass
    try:
        check_call(['kubectl', 'apply', '-f', './templates/loadbalancer.yaml',
                    '--namespace', cr.namespace, '--context', cr.context])
    except:
        pass

# Create/grab the load balancer IP(s)
for cr in create_clusters:
    external_ip = ''
    while True:
        external_ip = check_output(['kubectl', 'get', 'svc', 'cockroachdb-public', '--namespace', cr.namespace,
                                    '--context', cr.context, '--template', '{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}'])
        if external_ip:
            break
        print 'Waiting for load balancer IP in %s...' % (cr.namespace)
        sleep(10)
    print 'LB endpoint for namespace %s: %s ' % (cr.namespace, external_ip)
    cr.lb_ip = external_ip

# Build CA certs file

for cr in create_clusters:
    try:
        check_call('rm -r %s' % (cr.ca_certs_dir), shell=True)
    except:
        pass
    os.mkdir(cr.ca_certs_dir)
    check_call(['cockroach', 'cert', 'create-ca', '--certs-dir',
                cr.ca_certs_dir, '--ca-key', cr.ca_certs_dir+'/ca.key'])

for cr in create_clusters:
    for cr_join in create_clusters + join_clusters:
        if cr == cr_join:
            continue
        check_call(['cat %s >> %s' %
                    (cr_join.ca_certs_file, cr.ca_certs_file)], shell=True)


# Now we can set up the certs since we can get the lb's ip address.

# Build node and client certs
for cr in create_clusters:
    try:
        check_call('rm -r %s' % (cr.node_certs_dir), shell=True)
    except:
        pass
    try:
        check_call('rm -r %s' % (cr.client_certs_dir), shell=True)
    except:
        pass
    os.mkdir(cr.client_certs_dir)
    os.mkdir(cr.node_certs_dir)
    check_call(['cp', cr.ca_certs_file, cr.client_certs_dir])
    check_call(['cockroach', 'cert', 'create-client', 'root', '--certs-dir',
                cr.client_certs_dir, '--ca-key', cr.ca_certs_dir+'/ca.key'])

    check_call(['cp %s %s ' % (cr.client_certs_dir +
                               '/*', cr.node_certs_dir)], shell=True)

    check_call(['cockroach', 'cert', 'create-node', '--certs-dir', cr.node_certs_dir, '--ca-key', cr.ca_certs_dir+'/ca.key', cr.lb_ip, 'localhost', '127.0.0.1', 'cockroachdb-public', 'cockroachdb-public.default',
                'cockroachdb-public.'+cr.namespace, 'cockroachdb-public.%s.svc.cluster.local' % (cr.namespace), '*.cockroachdb', '*.cockroachdb.'+cr.namespace, '*.cockroachdb.%s.svc.cluster.local' % (cr.namespace)])
