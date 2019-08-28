#!/usr/bin/env python

import argparse
import itertools
import glob
import os
import shutil
import subprocess


class CockroachCluster(object):

    def __init__(self, namespace):
        self.namespace = namespace

    @property
    def directory(self):
        return os.path.join('generated', self.namespace)

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


def parse_args():
    parser = argparse.ArgumentParser(
        description='Creates certificates for a new Cockroachdb cluster')
    parser.add_argument('namespace', metavar='NAMESPACE')
    parser.add_argument('--node-address', metavar='ADDRESS', nargs='*',
        default=[], help='extra addresses to add to the node certificate')
    parser.add_argument('--node-ca-cert', metavar='FILENAME', nargs='*',
        default=[], help='paths to CA certificates of other clusters that the '
        'new cluster will join')

    return parser.parse_args()


def main():
    args = parse_args()
    cr = CockroachCluster(args.namespace)

    # Create the generated directories.
    try:
        os.mkdir('generated')
    except OSError:
        pass
    try:
        os.mkdir(cr.directory)
    except OSError:
        pass

    # Create a CA for each new cluster.
    # Delete and recreate the ca_certs_dir.
    shutil.rmtree(cr.ca_certs_dir, ignore_errors=True)
    os.mkdir(cr.ca_certs_dir)

    # Create the CA.
    subprocess.check_call([
        'cockroach', 'cert', 'create-ca',
        '--certs-dir', cr.ca_certs_dir,
        '--ca-key', os.path.join(cr.ca_certs_dir, 'ca.key')])

    # Combine the ca_certs_files of all the joined clusters to the new cluster's
    # ca_certs_file.
    with open(cr.ca_certs_file) as new_certs_fh:
        for ca_cert_file in args.node_ca_cert:
            with open(ca_cert_file) as ca_cert_fh:
                new_certs_fh.write(ca_cert_fh.read())
                new_certs_fh.write('\n')

    # Build node and client certs.
    # Delete and recreate the directories.
    shutil.rmtree(cr.node_certs_dir, ignore_errors=True)
    shutil.rmtree(cr.client_certs_dir, ignore_errors=True)
    os.mkdir(cr.client_certs_dir)
    os.mkdir(cr.node_certs_dir)

    shutil.copy(cr.ca_certs_file, cr.client_certs_dir)

    subprocess.check_call([
        'cockroach', 'cert', 'create-client', 'root',
        '--certs-dir', cr.client_certs_dir,
        '--ca-key', os.path.join(cr.ca_certs_dir, 'ca.key')])

    for filename in glob.glob(os.path.join(cr.client_certs_dir, '*')):
        shutil.copy(filename, cr.node_certs_dir)

    node_addresses = ['localhost']
    node_addresses.extend(args.node_address)
    node_addresses.extend([
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
        '--ca-key', os.path.join(cr.ca_certs_dir, 'ca.key')]
        + node_addresses)


if __name__ == '__main__':
    main()
