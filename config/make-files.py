#!/usr/bin/env python

import distutils.spawn
import json
import os
import stat
from subprocess import check_call,check_output
from sys import exit
from time import sleep


class CockroachCluster():

  def __init__(self, zone, context='', lb_ip='', ca_certs_file=''):
    self.zone = zone
    self.context = context
    self.lb_ip = lb_ip
    if ca_certs_file:
      self.ca_certs_file = ca_certs_file

  @property
  def directory(self):
    return os.path.join('./generated/', self.zone)

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

  def join_str(self):
    addrs = ['cockroachdb-%d.cockroachdb.%s' % (i, zone) for i in range(3)]
    for cr in create_clusters + join_clusters:
      if cr == self:
        continue
      addrs.append(cr.lb_ip)
    return ','.join(addrs)        


create_clusters = [
  # CockroachCluster(
  #   zone='us-central1-a',
  #   context='gke_wing-crdb-3_us-central1-a_cockroachdb3',
  # ),
]

join_clusters = [
  # CockroachCluster(
  #   lb_ip='external_ip_address',
  #   ca_certs_file='path_to_ca_public_cert',
  # ),
]

# Set up the necessary directories and certificates. Ignore errors because they may already exist.
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
    check_call(['kubectl', 'create', 'namespace', cr.zone, '--context', cr.context])
  except:
    pass
  try:
    check_call(['kubectl', 'apply', '-f', './templates/loadbalancer.yaml', '--namespace', cr.zone, '--context', cr.context])
  except:
    pass

for cr in create_clusters:
    external_ip = ''
    while True:
        external_ip = check_output(['kubectl', 'get', 'svc', 'cockroachdb-public', '--namespace', cr.zone, '--context', cr.context, '--template', '{{range .status.loadBalancer.ingress}}{{.ip}}{{end}}'])
        if external_ip:
            break
        print  'Waiting for load balancer IP in %s...' % (cr.zone)
        sleep(10)
    print 'LB endpoint for zone %s: %s' % (cr.zone, external_ip)
    cr.lb_ip = external_ip
  


# Create the cockroach resources in each cluster.
for cr in create_clusters:
  zone = cr.zone
  locality = 'zone=%s' % (zone)
  yaml_file = '%s/cockroachdb-statefulset-secure.yaml' % (cr.directory)
  with open(yaml_file, 'w') as f:
    check_call(['sed', 's/JOINLIST/%s/g;s/LOCALITYLIST/%s/g;s/PUBLIC_ADDR/%s/g' % (cr.join_str(), locality, cr.lb_ip), 'templates/cockroachdb-statefulset-secure.yaml'], stdout=f)

for cr in create_clusters:
  yaml_file = '%s/http-gateway.yaml' % (cr.directory)
  with open(yaml_file, 'w') as f:
    check_call(['sed', 's/YOUR_ZONE_HERE/%s/g' % (cr.zone), 'templates/http-gateway.yaml'], stdout=f)

# Copy the setup script into each directory.
cluster_init = 'true'
for cr in create_clusters:
  # Don't initialize the cluster if we're joining an existing one.
  if len(join_clusters) > 0:
    cluster_init = 'false'
  filename = '%s/setup.sh' % (cr.directory)
  with open(filename, 'w') as f:
    check_call(['sed', 's/REPLACE_CONTEXT/%s/g;s/REPLACE_NODE_CERTS_DIR/%s/g;s/REPLACE_CLIENT_CERTS_DIR/%s/g;s/REPLACE_ZONE/%s/g;s/REPLACE_DIRECTORY/%s/g;s/REPLACE_CLUSTER_INIT/%s/g' % (cr.context, cr.node_certs_dir.replace('/', '\/'), cr.client_certs_dir.replace('/', '\/'), cr.zone, cr.directory.replace('/', '\/'), cluster_init), 'templates/setup.sh'], stdout=f)
  # Only initialize one cluster.
  cluster_init = 'false'
  st = os.stat(filename)
  os.chmod(filename, st.st_mode | stat.S_IEXEC)

# Create a cockroachdb-public service in the default namespace in each cluster.
for cr in create_clusters:
  yaml_file = '%s/external-name-svc.yaml' % (cr.directory)
  with open(yaml_file, 'w') as f:
    check_call(['sed', 's/YOUR_ZONE_HERE/%s/g' % (cr.zone), 'templates/external-name-svc.yaml'], stdout=f)

for cr in create_clusters:
  try:
    check_call('rm -r %s' % (cr.ca_certs_dir), shell=True)
  except:
    pass
  os.mkdir(cr.ca_certs_dir)
  check_call(['cockroach', 'cert', 'create-ca', '--certs-dir', cr.ca_certs_dir, '--ca-key', cr.ca_certs_dir+'/ca.key'])

for cr in create_clusters:
  for cr_join in create_clusters + join_clusters:
    if cr == cr_join:
      continue
    check_call(['cat %s >> %s' % (cr_join.ca_certs_file, cr.ca_certs_file)], shell=True)


# Now we can set up the certs since we can get the lb's ip address.
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
  check_call(['cockroach', 'cert', 'create-client', 'root', '--certs-dir', cr.client_certs_dir, '--ca-key', cr.ca_certs_dir+'/ca.key'])

  check_call(['cp %s %s ' % (cr.client_certs_dir + '/*', cr.node_certs_dir)], shell=True)

  check_call(['cockroach', 'cert', 'create-node', '--certs-dir', cr.node_certs_dir, '--ca-key', cr.ca_certs_dir+'/ca.key', cr.lb_ip, 'localhost', '127.0.0.1', 'cockroachdb-public', 'cockroachdb-public.default', 'cockroachdb-public.'+cr.zone, 'cockroachdb-public.%s.svc.cluster.local' % (cr.zone), '*.cockroachdb', '*.cockroachdb.'+cr.zone, '*.cockroachdb.%s.svc.cluster.local' % (cr.zone)])
