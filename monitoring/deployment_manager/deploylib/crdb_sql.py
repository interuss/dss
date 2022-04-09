from dataclasses import dataclass
import datetime
import hashlib
import random
from typing import List, Tuple

from kubernetes import client as k8s
import kubernetes.stream


@dataclass
class User(object):
    username: str
    options: List[str]
    member_of: List[str]


def execute_sql(client: k8s.CoreV1Api, namespace: str, sql_commands: List[str]) -> str:
    """Execute the specfied sql_commands directly on a CRDB node.

    :param client: Kubernetes core client which can access the CRDB node
    :param namespace: Namespace in which an accessible CRDB node is located
    :param sql_commands: SQL commands to be executed (no semicolons)
    :return: stdout of console from executing `cockroach sql`
    """
    exec_command = ['./cockroach', 'sql', '--certs-dir=cockroach-certs/']
    exec_command += ['--execute=' + cmd for cmd in sql_commands]
    resp = kubernetes.stream.stream(client.connect_get_namespaced_pod_exec,
                  'cockroachdb-0',
                  namespace,
                  command=exec_command,
                  stderr=True, stdin=False,
                  stdout=True, tty=False)
    if 'Failed running' in resp:
        raise ValueError('SQL error: ' + resp)
    return resp


def list_users(client: k8s.CoreV1Api, namespace: str) -> List[User]:
    lines = execute_sql(client, namespace, ['SHOW USERS']).split('\n')
    lines = [line for line in lines[1:] if line]
    users = []
    for line in lines:
        username, options_text, member_of_text = [col.strip() for col in line.split('\t')]
        options = [opt.strip() for opt in options_text.split(',')]
        member_of = [group.strip() for group in member_of_text[1:-1].split(',')]
        users.append(User(username=username, options=options, member_of=member_of))
    return users


def get_monitoring_user(client: k8s.CoreV1Api, namespace: str, cluster_name: str) -> Tuple[str, str]:
    """Get the username and password for a CRDB user intended for monitoring.

    Whenever this routine is called, it sets the validity of the user to 1-2
    days beyond the current time.  To continue to use the user after that time,
    this routine must be called again.

    :param client: Kubernetes core client which can access the CRDB node
    :param namespace: Namespace in which an accessible CRDB node is located
    :param cluster_name: Name of Kubernetes cluster in which this user will be operating
    :return:
        * Username
        * Password
    """
    # Create username according to cluster_name
    prefix = 'monitoring_user_'
    suffix = hashlib.md5(cluster_name.encode('utf-8')).hexdigest()[-8:]
    username = prefix + suffix

    # Create a new password
    r = random.Random()
    password = ''.join(random.choice('abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-=_+[]\\{}|;:",./<>?') for _ in range(32))

    # Compute validity
    valid_until = (datetime.datetime.utcnow() + datetime.timedelta(days=2)).isoformat()
    execute_sql(
        client, namespace,
        ['CREATE USER IF NOT EXISTS {}'.format(username),
         'GRANT admin TO {}'.format(username),
         'ALTER USER {} WITH LOGIN PASSWORD \'{}\' VALID UNTIL \'{}\''.format(username, password, valid_until)])

    return username, password
