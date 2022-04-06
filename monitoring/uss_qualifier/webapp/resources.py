import rq
from redis import Redis
from monitoring.uss_qualifier.webapp import webapp
from . import config

redis_conn = Redis.from_url(webapp.config.get(config.KEY_REDIS_URL))


REDIS_KEY_UPLOADED_KMLS = 'uploaded_kmls'
REDIS_KEY_TEST_RUNS = 'test_runs'


qualifier_queue = rq.Queue(
    webapp.config.get(config.KEY_REDIS_QUEUE),
    connection=redis_conn,
    default_timeout=3600)
