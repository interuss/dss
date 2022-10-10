import math
import requests
from typing import Dict, List, Optional, Tuple

import arrow

from implicitdict import ImplicitDict


class Address(ImplicitDict):
    network_field: str
    address_field: str


class ServerVersion(ImplicitDict):
    major_val: int
    minor_val: int
    patch: int
    internal: int


class Locality(ImplicitDict):
    tiers: List[dict]


class Node(ImplicitDict):
    node_id: int
    address: Address
    locality: Locality
    ServerVersion: ServerVersion
    build_tag: str
    started_at: int
    cluster_name: str
    sql_address: Address
    metrics: dict
    total_system_memory: int
    num_cpus: int
    updated_at: int
    liveness_status: int

    def summarize(self) -> Tuple[str, Dict[str, str]]:
        key = 'Node {} ({})'.format(self.node_id, self.address.address_field)
        t0 = arrow.get(math.floor(self.started_at / 1e9))
        values = {
            'locality': ' '.join('{}:{}'.format(t['key'], t['value']) for t in self.locality.tiers),
            'status': 'Running {} since {}, liveness {}'.format(self.build_tag, t0.to('local').isoformat(),self.liveness_status)
        }
        return key, values


class ClusterAPI(object):
    """Wrapper for retrieving CockroachDB cluster information.

    API: https://www.cockroachlabs.com/docs/api/cluster/v2
    """
    def __init__(self, session: requests.Session, base_url: str='https://localhost:8080/api/v2', username: Optional[str]=None, password: Optional[str]=None):
        self._session = session
        self._base_url = base_url
        self._username = username
        self._password = password
        self._session_auth = None

    def __del__(self):
        self.log_out()

    def is_ready(self) -> bool:
        resp = self._session.get('{}/health/?ready=true'.format(self._base_url))
        if resp.status_code == 200:
            return True
        elif resp.status_code == 500:
            return False
        else:
            raise ValueError('Call to {} returned unexpected status code {}: {}'.format(resp.url, resp.status_code, resp.content.decode('utf-8')))

    def is_up(self) -> bool:
        resp = self._session.get('{}/health/'.format(self._base_url))
        if resp.status_code == 200:
            return True
        elif resp.status_code == 500:
            return False
        else:
            raise ValueError('Call to {} returned unexpected status code {}: {}'.format(resp.url, resp.status_code, resp.content.decode('utf-8')))

    def log_in(self) -> str:
        if self._username is None:
            raise ValueError('Cannot log in to CockroachDB cluster at {} when username is not specified'.format(self._base_url))
        if self._password is None:
            raise ValueError('Cannot log in to CockroachDB cluster at {} when password is not specified'.format(self._base_url))
        resp = self._session.post('{}/login/'.format(self._base_url), data={'username': self._username, 'password': self._password})
        resp.raise_for_status()
        session = resp.json().get('session', None)
        if session is None:
            raise ValueError('Invalid CockroachDB cluster response: `session` not specified: {}'.format(resp.content.decode('utf-8')))
        self._session_auth = session
        return session

    def _get_headers(self) -> Dict[str, str]:
        if self._session_auth is None:
            self.log_in()
        return {'X-Cockroach-API-Session': self._session_auth}

    def log_out(self) -> None:
        if self._session_auth is None:
            return
        resp = self._session.post('{}/logout/'.format(self._base_url), headers=self._get_headers())
        resp.raise_for_status()
        self._session_auth = None

    def get_nodes(self) -> List[Node]:
        resp = self._session.get('{}/nodes/'.format(self._base_url), headers=self._get_headers())
        resp.raise_for_status()
        nodes = resp.json().get('nodes', None)
        if nodes is None:
            raise ValueError('Invalid CockroachDB cluster response: `nodes` not specified: {}'.format(resp.content.decode('utf-8')))
        return [ImplicitDict.parse(n, Node) for n in nodes]
