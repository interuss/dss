import atexit
from datetime import timedelta, datetime
from multiprocessing import Process
import os
import time
from typing import Tuple, Optional

from loguru import logger
import requests
from uas_standards.interuss.automated_testing.flight_planning.v1.api import (
    ClearAreaRequest,
)

from implicitdict import ImplicitDict
from monitoring import mock_uss
from monitoring.atproxy.requests import (
    RequestType,
    SCDInjectionPutFlightRequest,
    SCDInjectionDeleteFlightRequest,
    SCDInjectionClearAreaRequest,
    SCD_REQUESTS,
)
from monitoring.atproxy.handling import (
    ListQueriesResponse,
    PutQueryRequest,
    PendingRequest,
)
from monitoring.mock_uss import config
from monitoring.mock_uss.scdsc.routes_injection import (
    injection_status,
    scd_capabilities,
    inject_flight,
    delete_flight,
    clear_area,
)
from monitoring.monitorlib.scd_automated_testing.scd_injection_api import (
    InjectFlightRequest,
)
from .database import db, Database, ATProxyWorker, ATProxyWorkerState, ATProxyWorkerID

MAX_DAEMON_PROCESSES = 1
ATPROXY_WAIT_TIMEOUT = timedelta(minutes=5)


def start() -> None:
    """Spawn daemon process to poll atproxy and fulfill its pending requests"""

    with db as tx:
        assert isinstance(tx, Database)
        if len(tx.atproxy_workers) < MAX_DAEMON_PROCESSES:
            new_worker_id = len(tx.atproxy_workers) + 1
            logger.info(
                "Starting atproxy client worker {} from process ID {}",
                new_worker_id,
                os.getpid(),
            )
            tx.atproxy_workers[new_worker_id] = ATProxyWorker(pid=os.getpid())
        else:
            logger.info(
                "No atproxy client worker needed from process ID {}", os.getpid()
            )
            new_worker_id = None

    if new_worker_id:
        p = Process(target=_atproxy_client_worker, args=(new_worker_id,))
        p.start()
        atexit.register(lambda: _stop(new_worker_id, p))


def _stop(worker_id: ATProxyWorkerID, p: Process) -> None:
    """Gracefully stop an atproxy handler daemon worker process

    Args:
        worker_id: atproxy worker ID of daemon process being stopped
        p: Process running daemon worker
    """
    with db as tx:
        assert isinstance(tx, Database)
        if tx.atproxy_workers[worker_id].pid != os.getpid():
            return
        if tx.atproxy_workers[worker_id].state in (
            ATProxyWorkerState.Running,
            ATProxyWorkerState.Starting,
        ):
            logger.info(
                "Requesting atproxy client worker {} to stop from process ID {}",
                worker_id,
                os.getpid(),
            )
            tx.atproxy_workers[worker_id].state = ATProxyWorkerState.Stopping
        else:
            logger.info(
                "atproxy client worker {} {} when stop requested from process ID {}",
                worker_id,
                tx.atproxy_workers[worker_id].state,
                os.getpid(),
            )
    p.join()
    logger.info(
        "Successfully stopped atproxy client worker {} from process ID {}",
        worker_id,
        os.getpid(),
    )


def _atproxy_client_worker(worker_id: ATProxyWorkerID) -> None:
    """Parent routine for an atproxy-polling daemon worker

    Args:
        worker_id: atproxy worker ID of this daemon worker
    """
    try:
        # Collect configuration information for this worker
        base_url = mock_uss.webapp.config[config.KEY_ATPROXY_BASE_URL]
        basic_auth_setting = mock_uss.webapp.config[config.KEY_ATPROXY_BASIC_AUTH]
        auth_components = tuple(s.strip() for s in basic_auth_setting.split(":"))
        if len(auth_components) != 2:
            raise ValueError(
                f'Invalid {config.ENV_KEY_ATPROXY_BASIC_AUTH}; expected <username>:<password> but instead found "{basic_auth_setting}"'
            )
        basic_auth = (auth_components[0], auth_components[1])

        # Start worker by making sure atproxy is reachable
        with db as tx:
            assert isinstance(tx, Database)
            tx.atproxy_workers[worker_id].state = ATProxyWorkerState.Starting
        _wait_for_atproxy(worker_id, base_url, basic_auth)

        # Mark worker as Running
        with db as tx:
            assert isinstance(tx, Database)
            if tx.atproxy_workers[worker_id].state != ATProxyWorkerState.Starting:
                logger.info(
                    "atproxy client worker {} from process ID {} will not start running because it is {}",
                    worker_id,
                    os.getpid(),
                    tx.atproxy_workers[worker_id].state,
                )
                return
            tx.atproxy_workers[worker_id].state = ATProxyWorkerState.Running

        # Enter the polling loop
        _poll_atproxy(worker_id, base_url, basic_auth)
    except Exception as e:
        logger.error(
            "atproxy client worker {} from process ID {} encountered critical error: {}",
            worker_id,
            os.getpid(),
            str(e),
        )
        raise
    finally:
        # Gracefully stop worker
        logger.info(
            "atproxy client worker {} from process ID {} is stopping",
            worker_id,
            os.getpid(),
        )
        with db as tx:
            assert isinstance(tx, Database)
            tx.atproxy_workers[worker_id].state = ATProxyWorkerState.Stopped


def _wait_for_atproxy(
    worker_id: ATProxyWorkerID, base_url: str, basic_auth: Tuple[str, str]
) -> None:
    """Wait for atproxy to be available

    Args:
        worker_id: atproxy worker ID of this daemon worker
        base_url: base URL of remote atproxy instance
        basic_auth: (username, password) tuple for accessing atproxy
    """
    timeout = datetime.utcnow() + ATPROXY_WAIT_TIMEOUT
    status_url = f"{base_url}/status"
    while db.value.atproxy_workers[worker_id].state == ATProxyWorkerState.Starting:
        resp = None
        try:
            resp = requests.get(status_url, auth=basic_auth)
            if resp.status_code == 200:
                break
            logger.info(
                "atproxy at {} is not yet ready; received {} at /status: {}",
                base_url,
                resp.status_code,
                resp.content.decode(),
            )
        except requests.exceptions.ConnectionError as e:
            logger.info("atproxy at {} is not yet reachable: {}", base_url, str(e))
        if datetime.utcnow() > timeout:
            raise RuntimeError(
                f"Timeout while trying to connect to atproxy at {status_url}; latest attempt yielded {resp.status_code if resp else 'ConnectionError'}"
            )
        time.sleep(5)


def _poll_atproxy(
    worker_id: ATProxyWorkerID, base_url: str, basic_auth: Tuple[str, str]
) -> None:
    """Poll atproxy for new requests and handle any unhandled requests

    Args:
        worker_id: atproxy worker ID of this daemon worker
        base_url: base URL of remote atproxy instance
        basic_auth: (username, password) tuple for accessing atproxy
    """
    query_url = f"{base_url}/handler/queries"
    while db.value.atproxy_workers[worker_id].state == ATProxyWorkerState.Running:
        # Poll atproxy to see if there are any requests pending
        resp = requests.get(query_url, auth=basic_auth)
        if resp.status_code != 200:
            logger.error(
                "Error {} polling {}: {}",
                resp.status_code,
                query_url,
                resp.content.decode(),
            )
            time.sleep(5)
            continue
        try:
            queries_resp: ListQueriesResponse = ImplicitDict.parse(
                resp.json(), ListQueriesResponse
            )
        except ValueError as e:
            logger.error(
                "Error parsing atproxy response to request for queries: {}", str(e)
            )
            time.sleep(5)
            continue
        if not queries_resp.requests:
            logger.debug("No queries currently pending.")
            continue

        # Identify a request to handle
        request_to_handle = None
        with db as tx:
            for req in queries_resp.requests:
                if req.id not in {
                    w.handling_request for w in tx.atproxy_workers.values()
                }:
                    request_to_handle = req
                    tx.atproxy_workers[worker_id].handling_request = req.id
                    break
        if not request_to_handle:
            continue

        # Handle the request
        logger.info("Handling response to {} request", request_to_handle.type)
        fulfillment = PutQueryRequest(
            return_code=500,
            response={"message": "Unknown error in mock_uss atproxy client handler"},
        )
        try:
            content, code = _fulfill_request(request_to_handle)
            fulfillment = PutQueryRequest(return_code=code, response=content)
        except ValueError as e:
            msg = f"mock_uss atproxy client handler encountered ValueError: {e}"
            logger.error(msg)
            fulfillment = PutQueryRequest(
                return_code=400,
                response={"message": msg},
            )
        except NotImplementedError as e:
            msg = (
                f"mock_uss atproxy client handler encountered NotImplementedError: {e}"
            )
            logger.error(msg)
            fulfillment = PutQueryRequest(
                return_code=500,
                response={"message": msg},
            )
        finally:
            resp = requests.put(
                f"{query_url}/{request_to_handle.id}", json=fulfillment, auth=basic_auth
            )
            if resp.status_code != 200:
                logger.error(
                    f"Error {resp.status_code} reporting response {fulfillment.return_code} to query {request_to_handle.id}: {resp.content.decode()}"
                )


def _fulfill_request(request_to_handle: PendingRequest) -> Tuple[Optional[dict], int]:
    """Fulfill a PendingRequest from atproxy by invoking appropriate handler logic

    Args:
        request_to_handle: PendingRequest to handle

    Returns:
        * dict content of response, or None for no response JSON body
        * HTTP status code of response
    """
    req_type = request_to_handle.type

    if req_type in SCD_REQUESTS:
        if mock_uss.SERVICE_SCDSC not in mock_uss.enabled_services:
            raise ValueError(
                f"mock_uss cannot handle {req_type} request because {mock_uss.SERVICE_SCDSC} is not one of the enabled services ({', '.join(mock_uss.enabled_services)})"
            )

    if req_type == RequestType.SCD_GetStatus:
        return injection_status()
    elif req_type == RequestType.SCD_GetCapabilities:
        return scd_capabilities()
    elif req_type == RequestType.SCD_PutFlight:
        req = ImplicitDict.parse(
            request_to_handle.request, SCDInjectionPutFlightRequest
        )
        body = ImplicitDict.parse(req.request_body, InjectFlightRequest)
        return inject_flight(req.flight_id, body)
    elif req_type == RequestType.SCD_DeleteFlight:
        req = ImplicitDict.parse(
            request_to_handle.request, SCDInjectionDeleteFlightRequest
        )
        return delete_flight(req.flight_id)
    elif req_type == RequestType.SCD_CreateClearAreaRequest:
        req = ImplicitDict.parse(
            request_to_handle.request, SCDInjectionClearAreaRequest
        )
        body = ImplicitDict.parse(req.request_body, ClearAreaRequest)
        return clear_area(body)
    else:
        # TODO: Add RID injection & observation support
        raise NotImplementedError(f"Unsupported request type: {request_to_handle.type}")
