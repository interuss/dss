import rq
from redis import Redis
from monitoring.rid_qualifier.webapp import webapp
from . import config


qualifier_queue = rq.Queue(
    webapp.config.get(config.KEY_REDIS_QUEUE),
    connection=Redis.from_url(webapp.config.get(config.KEY_REDIS_URL)),
    default_timeout=3600)
