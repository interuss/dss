import rq
from redis import Redis
from monitoring.uss_qualifier.webapp import webapp
from . import config

redis_conn = Redis.from_url(webapp.config.get(config.KEY_REDIS_URL))


REDIS_KEY_UPLOADED_KMLS = 'uploaded_kmls'
REDIS_KEY_TEST_RUNS = 'test_runs'
REDIS_KEY_TEMP_LOGS = 'temp_logs'
REDIS_KEY_TEST_RUN_LOGS = 'test_run_logs'


qualifier_queue = rq.Queue(
    webapp.config.get(config.KEY_REDIS_QUEUE),
    connection=redis_conn,
    default_timeout=3600)

def decode_redis(src):
    if isinstance(src, list):
        rv = list()
        for key in src:
            rv.append(decode_redis(key))
        return rv
    elif isinstance(src, dict):
        rv = dict()
        for key in src:
            rv[key.decode()] = decode_redis(src[key])
        return rv
    elif isinstance(src, bytes):
        return src.decode()
    else:
        raise Exception("type not handled: " +type(src))
