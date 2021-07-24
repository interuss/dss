import json
import time
import redis
import rq
from . import config
from monitoring.rid_qualifier import test_executor
from monitoring.rid_qualifier.utils import RIDQualifierTestConfiguration
from monitoring.monitorlib.typing import ImplicitDict

def example(seconds):
    print('Starting task')
    for i in range(seconds):
        print(i)
        time.sleep(1)
    print('Task completed')
    return 'task completed'

def get_rq_job(job_id):
  try:
      rq_job = config.Config.qualifier_queue.fetch_job(job_id)
  except (redis.exceptions.RedisError, rq.exceptions.NoSuchJobError):
      return None
  return rq_job

# def call_test_executor(user_config_json, auth_spec):
#     user_config: RIDQualifierTestConfiguration = ImplicitDict.parse(
#         json.loads(user_config_json), RIDQualifierTestConfiguration)
#     job = config.Config.qualifier_queue.enqueue(test_executor.main, user_config, auth_spec)
#     return job.get_id()

def call_test_executor(user_config_json, auth_spec):
    user_config: RIDQualifierTestConfiguration = ImplicitDict.parse(
        json.loads(user_config_json), RIDQualifierTestConfiguration)
    test_executor.main(user_config, auth_spec)