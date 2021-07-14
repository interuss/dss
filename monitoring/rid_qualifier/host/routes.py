import flask
import json
import rq
import time

from . import config
from flask import render_template
from redis import Redis
from werkzeug.exceptions import HTTPException

from monitoring.monitorlib import versioning, auth_validation
from monitoring.rid_qualifier.host import webapp
from . import forms


@webapp.route('/')
def home_page():
  return render_template('home.html', title='Home', greetings='Hello RID Host !!')

@webapp.route('/start_task', methods=['GET', 'POST'])
def start_background_task():
  queue = rq.Queue(
    config.Config.REDIS_QUEUE,
    connection=Redis.from_url(config.Config.REDIS_URL))
  job = queue.enqueue('monitoring.rid_qualifier.host.tasks.example', 1)
  time.sleep(3)
  return json.dumps({'job_id': job.get_id(), 'job_finished': job.is_finished})

@webapp.route('/userconfig')
def login():
    form = forms.UserConfig()
    return render_template('config_form.html', title='Get User config', form=form)

@webapp.route('/status')
def status():
  return 'Mock Host Service Provider ok {}'.format(versioning.get_code_version())


@webapp.errorhandler(Exception)
def handle_exception(e):
  if isinstance(e, HTTPException):
    return e
  elif isinstance(e, auth_validation.InvalidScopeError):
    return flask.jsonify({
      'message': 'Invalid scope; expected one of {%s}, but received only {%s}' % (
        ' '.join(e.permitted_scopes),
        ' '.join(e.provided_scopes))}), 403
  elif isinstance(e, auth_validation.InvalidAccessTokenError):
    return flask.jsonify({'message': e.message}), 401
  elif isinstance(e, auth_validation.ConfigurationError):
    return flask.jsonify({'message': e.message}), 500
  elif isinstance(e, ValueError):
    return flask.jsonify({'message': str(e)}), 400

  return flask.jsonify({'message': str(e)}), 500
