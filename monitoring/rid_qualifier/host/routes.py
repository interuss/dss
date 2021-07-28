import flask

from . import config
from . import forms
from . import tasks

from flask import render_template, flash, request, make_response
from werkzeug.exceptions import HTTPException

from monitoring.monitorlib import versioning, auth_validation
from monitoring.rid_qualifier.host import webapp


@webapp.route('/')
def home_page():
  return render_template('home.html', title='Home', greetings='Hello RID Host !!')


def start_background_task(user_config, auth_spec, debug):
  job = config.Config.qualifier_queue.enqueue(
    'monitoring.rid_qualifier.host.tasks.call_test_executor', user_config, auth_spec, debug)
  return job.get_id()

@webapp.route('/userconfig', methods=['GET', 'POST'])
def user_config():
    form = forms.UserConfig()
    job_id = ''
    data = {}
    if form.validate_on_submit():
      job_id = start_background_task(
        form.user_config.data, form.auth_spec.data, form.sample_report.data)
    if request.method == 'POST':
      data = {
        'job_id' : job_id
      }
    return render_template('start_task.html', title='Get User config', form=form, data=data)

@webapp.route('/result/<string:job_id>', methods=['GET', 'POST'])
def get_result(job_id):
  task = tasks.get_rq_job(job_id)
  response_object = {}
  if task:
      response_object = {
          "task_id": task.get_id(),
          "task_status": task.get_status(),
          "task_result": task.result,
      }
  return response_object


@webapp.route('/report/<string:job_id>', methods=['POST'])
def get_report(job_id):
  response_object = config.Config.redis_client.get(job_id)
  output = make_response(response_object)
  output.headers["Content-Disposition"] = "attachment; filename=report.json"
  output.headers["Content-type"] = "text/csv"
  return output


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
