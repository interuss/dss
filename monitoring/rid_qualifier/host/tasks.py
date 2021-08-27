import json
import time
import redis
import rq
from . import config, resources
from monitoring.rid_qualifier.host import webapp
# import resources
from monitoring.monitorlib.typing import ImplicitDict
from monitoring.rid_qualifier import test_executor
from monitoring.rid_qualifier.utils import RIDQualifierTestConfiguration
from monitoring.rid_qualifier.test_data import test_report
# from resources import qualifier_queue


def get_rq_job(job_id):
  try:
      rq_job = resources.qualifier_queue.fetch_job(job_id)
  except (redis.exceptions.RedisError, rq.exceptions.NoSuchJobError):
      return None
  return rq_job


def call_test_executor(user_config_json, auth_spec, input_files, debug=False):
    user_config: RIDQualifierTestConfiguration = ImplicitDict.parse(
        json.loads(user_config_json), RIDQualifierTestConfiguration)
    if debug:
        report = test_report.test_data
    else:
        report = test_executor.main(user_config, auth_spec, input_files)
    return json.dumps(report)
