from monitoring.uss_qualifier.test_data import test_report
from monitoring.uss_qualifier.utils import USSQualifierTestConfiguration
from monitoring.uss_qualifier.main import uss_test_executor
from monitoring.uss_qualifier.rid.simulator import flight_state_from_kml
from monitoring.uss_qualifier.rid.utils import FullFlightRecord
from monitoring.uss_qualifier.rid.utils import FullFlightRecord
import json
from typing import List
import redis
import rq
import uuid
from . import resources
from implicitdict import ImplicitDict


def get_rq_job(job_id):
    try:
        rq_job = resources.qualifier_queue.fetch_job(job_id)
    except (redis.exceptions.RedisError, rq.exceptions.NoSuchJobError):
        return None
    return rq_job


def remove_rq_job(job_id):
    """Removes a job from the queue."""
    try:
        rq_job = resources.qualifier_queue.remove(job_id)
    except (redis.exceptions.RedisError, rq.exceptions.NoSuchJobError):
        return None
    return rq_job


def call_test_executor(
    user_config_json: str,
    auth_spec: str,
    flight_record_jsons: List[str],
    testruns_id,
    debug=False,
    scd_test_definitions_path=None,
):

    config_json = json.loads(user_config_json)
    config: USSQualifierTestConfiguration = ImplicitDict.parse(
        config_json, USSQualifierTestConfiguration
    )
    flight_records: List[FullFlightRecord] = [
        ImplicitDict.parse(json.loads(j), FullFlightRecord) for j in flight_record_jsons
    ]
    if debug:
        report = json.dumps(test_report.test_data)
    else:
        report = json.dumps(
            uss_test_executor(
                config, auth_spec, flight_records, scd_test_definitions_path
            )
        )
    resources.redis_conn.hset(resources.REDIS_KEY_TEST_RUNS, testruns_id, report)
    return report


def call_kml_processor(kml_content, output_path):
    flight_states = flight_state_from_kml.main(
        kml_content, output_path, from_string=True
    )
    resources.redis_conn.hset(
        resources.REDIS_KEY_UPLOADED_KMLS, str(uuid.uuid4()), json.dumps(flight_states)
    )
    return flight_states
