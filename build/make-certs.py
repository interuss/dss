#!/usr/bin/env python3

import argparse
import itertools
import glob
import os
import shutil
import subprocess


class CockroachCluster(object):

    def __init__(self, cluster_context, namespace, ca_cert_to_join=None):
        self._ca_cert_to_join = ca_cert_to_join
        self._cluster_context = cluster_context
        self._namespace = namespace

    @property
    def ca_cert_to_join(self):
        return self._ca_cert_to_join

    @property
    def namespace(self):
        return self._namespace

    @property
    def directory(self):
        return os.path.join('workspace', self._cluster_context)

    @property
    def ca_certs_file(self):
        return os.path.join(self.ca_certs_dir, 'ca.crt')

    @property
    def ca_key_dir(self):
        return os.path.join(self.directory, 'ca_key_dir')

    @property
    def ca_key_file(self):
        return os.path.join(self.ca_key_dir, 'ca.key')

    @property
    def ca_certs_dir(self):
        return os.path.join(self.directory, 'ca_certs_dir')

    @property
    def client_certs_dir(self):
        return os.path.join(self.directory, 'client_certs_dir')

    @property
    def node_certs_dir(self):
        return os.path.join(self.directory, 'node_certs_dir')


def parse_args():
    parser = argparse.ArgumentParser(
        description='Creates certificates for a new Cockroachdb cluster')
    parser.add_argument('--cluster-context', metavar='CLUSTER_CONTEXT', required=True,
                        help='kubernetes cluster context name')
    parser.add_argument('--namespace', metavar='NAMESPACE', required=True,
                        help='kubernetes cluster namespace you are deploying to.')
    parser.add_argument('--node-address', metavar='ADDRESS', nargs='*',
                        default=[], help='extra addresses to add to the node certificate')
    parser.add_argument('--ca-cert-to-join', metavar='FILENAME',
                        help='file containing an existing CA cert of a cluster to join.')
    parser.add_argument('--overwrite-ca-cert', action='store_true', default=False,
                        help='True to generate new CA certs, false to use the existing one')
    return parser.parse_args()


def main():
    args = parse_args()
    cr = CockroachCluster(args.cluster_context,
                          args.namespace, args.ca_cert_to_join)

    # Create the generated directories.
    os.makedirs('workspace', exist_ok=True)
    os.makedirs(cr.directory, exist_ok=True)

    if args.overwrite_ca_cert:
      # Create a new CA.
      # Delete and recreate the ca_certs_dir.
      shutil.rmtree(cr.ca_certs_dir, ignore_errors=True)
      shutil.rmtree(cr.ca_key_dir, ignore_errors=True)
      os.mkdir(cr.ca_certs_dir)
      os.mkdir(cr.ca_key_dir)

    # Build node and client certs.
    # Delete and recreate the directories.
    shutil.rmtree(cr.node_certs_dir, ignore_errors=True)
    shutil.rmtree(cr.client_certs_dir, ignore_errors=True)
    os.mkdir(cr.client_certs_dir)
    os.mkdir(cr.node_certs_dir)

    if args.overwrite_ca_cert:
      # Create the CA.
      subprocess.check_call([
          'cockroach', 'cert', 'create-ca',
          '--certs-dir', cr.ca_certs_dir,
          '--ca-key', cr.ca_key_file])

    # Copy out the CA cert for generation, we delete these copies later.
    shutil.copy(cr.ca_certs_file, cr.client_certs_dir)
    shutil.copy(cr.ca_certs_file, cr.node_certs_dir)

    # We slightly abuse the rotate certs feature:
    # https://www.cockroachlabs.com/docs/stable/rotate-certificates.html
    if cr.ca_cert_to_join:
        with open(cr.ca_certs_file, 'a') as new_certs_fh:
            with open(cr.ca_cert_to_join) as join_ca_cert_fh:
                new_certs_fh.write(join_ca_cert_fh.read())
                new_certs_fh.write('\n')

    if cr.ca_cert_to_join and args.overwrite_ca_cert:
      print('Created new CA certificate in {}'.format(cr.ca_certs_dir))

    subprocess.check_call([
        'cockroach', 'cert', 'create-client', 'root',
        '--certs-dir', cr.client_certs_dir,
        '--ca-key', cr.ca_key_file])

    print('Created new client certificate in {}'.format(cr.client_certs_dir))

    node_addresses = ['localhost']
    node_addresses.extend(args.node_address)
    node_addresses.extend([
        'cockroachdb-balanced',
        'cockroachdb-balanced.%s' % cr.namespace,
        'cockroachdb-balanced.%s.svc.cluster.local' % cr.namespace,
        '*.cockroachdb',
        '*.cockroachdb.%s' % cr.namespace,
        'cockroachdb.%s' % cr.namespace,
        '*.cockroachdb.%s.svc.cluster.local' % cr.namespace
    ])

    subprocess.check_call([
        'cockroach', 'cert', 'create-node',
        '--certs-dir', cr.node_certs_dir,
        '--ca-key', cr.ca_key_file]
        + node_addresses)

    os.remove(os.path.join(cr.node_certs_dir, 'ca.crt'))
    os.remove(os.path.join(cr.client_certs_dir, 'ca.crt'))

    print('Created new node certificate in {}'.format(cr.node_certs_dir))


if __name__ == '__main__':
    main()
