from http.client import HTTPConnection
from urllib3 import PoolManager, HTTPConnectionPool, HTTPSConnectionPool
import urllib3
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)
import urllib3.connection
from typing import Tuple

from kubernetes.client import CoreV1Api
import kubernetes.stream
import requests
from requests.adapters import HTTPAdapter


def _get_session_for_socket(sock, host: str, port: int) -> requests.Session:
    sock.setblocking(True)
    s = requests.Session()
    s.verify = False
    http_adapter = SocketHTTPAdapter(sock, host, port)
    s.mount('http://', http_adapter)
    s.mount('https://', http_adapter)
    return s


def get_requests_session_for_pod(pod_name: str, namespace: str, port: int, client: CoreV1Api) -> Tuple[requests.Session, str]:
    """Make a Session to connect to the specified port on the named pod.

    Retrieves a requests Session that will send http(s) queries to the specified
    port on the specified pod.  Only requests sent to the returned host:port
    combination will be sent to the pod; other requests will result in errors.

    :param pod_name: Name of Kubernetes pod to connect to
    :param namespace: Kubernetes namespace in which the pod is located
    :param port: Pod's port to connect to
    :param client: Kubernetes client that can access the Kubernetes cluster
    :return:
        * requests.Session that will can be used to make HTTP(S) requests
        * host:port combination to which requests should be addressed
    """
    pf = kubernetes.stream.portforward(
        client.connect_get_namespaced_pod_portforward,
        pod_name, namespace, ports=str(port),
    )
    host = '{}.pod.{}'.format(pod_name, namespace)
    return _get_session_for_socket(pf.socket(port), host, port), '{}:{}'.format(host, port)


# ==== Custom requests handlers to inject/use an explicitly-provided socket ====


class SocketHTTPAdapter(HTTPAdapter):
    def __init__(self, sock, host: str, port: int, *args, **kwargs):
        self._socket = sock
        self._host = host
        self._port = port
        super(SocketHTTPAdapter, self).__init__(*args, **kwargs)

    def init_poolmanager(self, connections, maxsize, block=requests.adapters.DEFAULT_POOLBLOCK, **pool_kwargs):
        """Overrides method in base class"""
        self.poolmanager = SocketPoolManager(self._socket, self._host, self._port, num_pools=connections, maxsize=maxsize)


class SocketPoolManager(PoolManager):
    def __init__(self, sock, host: str, port: int, *args, **kwargs):
        self._socket = sock
        self._host = host
        self._port = port
        super(SocketPoolManager, self).__init__(*args, **kwargs)

    def _new_pool(self, scheme, host, port, request_context=None):
        """Overrides method in base class"""
        # return super(SocketPoolManager, self)._new_pool(scheme, host, port, request_context)
        if host == self._host and port == self._port:
            if scheme == 'http':
                return SocketHTTPConnectionPool(self._socket, host, port, **self.connection_pool_kw)
            elif scheme == 'https':
                return SocketHTTPSConnectionPool(self._socket, host, port, **self.connection_pool_kw)
        raise ValueError('{}:{} is not supported by SocketPoolManager intended for {}:{}'.format(host, port, self._host, self._port))


class SocketHTTPConnectionPool(HTTPConnectionPool):
    def __init__(self, sock, *args, **kwargs):
        self._socket = sock
        super(SocketHTTPConnectionPool, self).__init__(*args, **kwargs)

    def _new_conn(self):
        """Overrides method in base class"""
        self.num_connections += 1
        return SocketHTTPConnection(
            self._socket,
            host=self.host,
            port=self.port,
            timeout=self.timeout.connect_timeout,
            **self.conn_kw
        )


class SocketHTTPConnection(HTTPConnection):
    def __init__(self, sock, *args, **kwargs):
        self._socket = sock
        super(SocketHTTPConnection, self).__init__(*args, **kwargs)

    def connect(self):
        """Overrides method in base class"""
        self.sock = self._socket
        if self._tunnel_host:
            self._tunnel()


class SocketHTTPSConnectionPool(HTTPSConnectionPool):
    def __init__(self, sock, *args, **kwargs):
        self._socket = sock
        kwargs['assert_hostname'] = False
        super(SocketHTTPSConnectionPool, self).__init__(*args, **kwargs)

    def _new_conn(self):
        """Overrides method in base class"""
        self.num_connections += 1

        actual_host = self.host
        actual_port = self.port
        if self.proxy is not None:
            actual_host = self.proxy.host
            actual_port = self.proxy.port

        conn = SocketHTTPSConnection(
            self._socket,
            host=actual_host,
            port=actual_port,
            timeout=self.timeout.connect_timeout,
            cert_file=self.cert_file,
            key_file=self.key_file,
            **self.conn_kw
        )

        return self._prepare_conn(conn)


class SocketHTTPSConnection(urllib3.connection.HTTPSConnection):
    def __init__(self, sock, *args, **kwargs):
        self._socket = sock
        super(SocketHTTPSConnection, self).__init__(*args, **kwargs)

    def _new_conn(self):
        """Overrides method in base class"""
        return self._socket
