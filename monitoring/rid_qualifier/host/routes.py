import flask
import os

from . import config
from . import forms
from . import tasks

from flask import render_template, request, make_response, redirect, url_for
from werkzeug.exceptions import HTTPException
from werkzeug.utils import secure_filename

from monitoring.monitorlib import versioning, auth_validation
from monitoring.rid_qualifier.host import webapp

import os
import pathlib

import requests
from flask import Flask, session, abort, redirect, request
from google.oauth2 import id_token
from google_auth_oauthlib.flow import Flow
from pip._vendor import cachecontrol
import google.auth.transport.requests

os.environ["OAUTHLIB_INSECURE_TRANSPORT"] = "1"

GOOGLE_CLIENT_ID = "95537707830-qovbcpoe6akdmk2tmai2su2pomp2hu4k.apps.googleusercontent.com"
client_secrets_file = os.path.join(pathlib.Path(__file__).parent, "client_secret.json")

flow = Flow.from_client_secrets_file(
    client_secrets_file=client_secrets_file,
    scopes=["https://www.googleapis.com/auth/userinfo.profile", "https://www.googleapis.com/auth/userinfo.email", "openid"],
    redirect_uri="http://127.0.0.1:8072/callback"
)

def login_is_required(function):
    def wrapper(*args, **kwargs):
        if "google_id" not in session:
            return abort(401)  # Authorization required
        else:
            return function()

    return wrapper

@webapp.route("/login")
def login():
    authorization_url, state = flow.authorization_url()
    session["state"] = state
    return redirect(authorization_url)

@webapp.route("/callback")
def callback():
  # return request.url
    flow.fetch_token(authorization_response=request.url)

    if not session["state"] != request.args["state"]:
        abort(401)  # State does not match!

    credentials = flow.credentials
    request_session = requests.session()
    cached_session = cachecontrol.CacheControl(request_session)
    token_request = google.auth.transport.requests.Request(session=cached_session)

    id_info = id_token.verify_oauth2_token(
        id_token=credentials._id_token,
        request=token_request,
        audience=GOOGLE_CLIENT_ID
    )

    session["google_id"] = id_info.get("sub")
    session["name"] = id_info.get("name")
    return redirect("/")

@webapp.route('/')
def home_page():
  return render_template('home.html', title='Home', greetings='Hello RID Host !!')


def start_background_task(user_config, auth_spec, input_files, debug):
  job = config.Config.qualifier_queue.enqueue(
    'monitoring.rid_qualifier.host.tasks.call_test_executor',
    user_config, auth_spec, input_files, debug)
  return job.get_id()


@webapp.route('/executor', methods=['GET', 'POST'])
def execute_task():
    files = request.args['files']
    if not files:
      return 'files not found.'
    files = files.split(',')
    form = forms.UserConfig(file_count=len(files))
    job_id = ''
    data = {}
    if form.validate_on_submit():
      job_id = start_background_task(
        form.user_config.data, form.auth_spec.data, files, form.sample_report.data)
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


@webapp.route('/upload_file', methods=['POST'])
def upload_flight_state_files():
    """Upload files."""
    files = request.files.getlist('files[]')
    destination_file_paths = []
    for file in files:
        if file:
            filename = secure_filename(file.filename)
            if filename.endswith('.json'):
              file_path = os.path.join(config.Config.INPUT_PATH, filename)
              file.save(file_path)
              destination_file_paths.append(file_path)
    return redirect(url_for('.execute_task', files=','.join(destination_file_paths)))


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
