import flask
import json
import logging
import os
import pathlib
import requests

from . import config
from . import forms
from . import tasks

from datetime import datetime
from google.oauth2 import id_token
from google_auth_oauthlib.flow import Flow
import google.auth.transport.requests
from flask import render_template, request, make_response, redirect, url_for, session
from functools import wraps
from pip._vendor import cachecontrol
from werkzeug.exceptions import HTTPException
from werkzeug.utils import secure_filename

from monitoring.monitorlib import versioning, auth_validation
from monitoring.rid_qualifier.host import webapp

logging.basicConfig(level=logging.INFO)

client_secrets_file = os.path.join(
    pathlib.Path(__file__).parent,
    'client_secret.json')

try:
    flow = Flow.from_client_secrets_file(
        client_secrets_file=client_secrets_file,
        scopes=[
            'https://www.googleapis.com/auth/userinfo.profile',
            'https://www.googleapis.com/auth/userinfo.email',
            'openid'],
        redirect_uri=f'{config.ENV_KEY_RID_HOST_URL}/login_callback')
except FileNotFoundError:
    flow = ''


@webapp.route('/info')
def info():
    return render_template('info.html')

def login_required(function):
    @wraps(function)
    def decorated_function(*args, **kwargs):
        if flow and 'google_id' not in session:
            return redirect(url_for('login', next=request.url))
        elif 'google_id' not in session:
            session['google_id'] = 'testuser'
            session['name'] = 'Test User'
            session['state'] = 'test'
        return function(*args, **kwargs)
    return decorated_function


@webapp.route('/login')
def login():
    # make sure session is empty
    session.clear()
    if not flow:
        return redirect(url_for('.info'))
    authorization_url, state = flow.authorization_url()
    session['state'] = state
    return redirect(authorization_url)


@webapp.route('/login_callback')
def login_callback():
    if not flow:
        return redirect(url_for('.info'))
    flow.fetch_token(authorization_response=request.url)

    credentials = flow.credentials
    request_session = requests.session()
    cached_session = cachecontrol.CacheControl(request_session)
    token_request = google.auth.transport.requests.Request(
        session=cached_session)

    id_info = id_token.verify_oauth2_token(
        id_token=credentials._id_token,
        request=token_request
    )

    session['google_id'] = id_info.get('sub')
    session['name'] = id_info.get('name')
    return redirect('/')


@webapp.route('/logout')
def logout():
    session.clear()
    return render_template('logout.html')


def _start_background_task(user_config, auth_spec, input_files, debug):
    job = config.Config.qualifier_queue.enqueue(
        'monitoring.rid_qualifier.host.tasks.call_test_executor',
        user_config, auth_spec, input_files, debug)
    return job.get_id()

def _get_running_jobs():
    registry = config.Config.qualifier_queue.started_job_registry
    running_job = registry.get_job_ids()
    if running_job:
        return running_job[0]

def _process_kml_files_task(kml_content, output_path):
    job = config.Config.qualifier_queue.enqueue(
        'monitoring.rid_qualifier.host.tasks.call_kml_processor',
        kml_content, output_path)
    return job.get_id()

@webapp.route('/', methods=['GET', 'POST'])
@webapp.route('/tests', methods=['GET', 'POST'])
@login_required
def tests():
    files = []
    running_job = _get_running_jobs()
    flight_record_data = get_flight_records()

    if flight_record_data.get('flight_records'):
        files= [(x, x) for x in flight_record_data['flight_records']]

    form = forms.TestsExecuteForm()
    form.flight_records.choices = files
    data = get_test_history()
    if running_job:
        data.update({'job_id': running_job})
    else:
        job_id = ''
        if form.validate_on_submit():
            file_objs = []
            user_id = session['google_id']
            input_files_location = f'{config.Config.FILE_PATH}/{user_id}/flight_records'
            for filename in form.flight_records.data:
                filepath = f'{input_files_location}/{filename}'
                with open(filepath) as fo:
                    file_objs.append(fo.read())
                job_id = _start_background_task(
                    form.user_config.data,
                    form.auth_spec.data,
                    file_objs,
                    form.sample_report.data)
            if request.method == 'POST':
                data.update({
                    'job_id': job_id,
                })
    return render_template(
        'tests.html',
        title='Execute tests',
        form=form,
        data=data)


def get_flight_records():
    data = {
        'flight_records': [],
        'message': ''
    }
    user_id = session['google_id']
    folder_path = f'{config.Config.FILE_PATH}/{user_id}/flight_records'
    if not os.path.isdir(folder_path):
        data['message'] = 'Flight records not available.'
    else:
        flight_records = [f for f in os.listdir(folder_path) if f.endswith('.json')]
        data['flight_records'] = flight_records
    return data

def _write_to_file(filepath, content):
    os.makedirs(os.path.dirname(filepath), exist_ok=True)
    with open(filepath, 'w') as f:
        f.write(content)

@webapp.route('/result/<string:job_id>', methods=['GET', 'POST'])
def get_result(job_id):
    task = tasks.get_rq_job(job_id)
    response_object = {}
    if task:
        response_object = {
            'task_id': task.get_id(),
            'task_status': task.get_status(),
            'task_result': task.result,
        }
    if task.get_status() == 'finished':
        task_result = task.result
        response_object.update({
            'task_status': 'finished',
            'task_result': task_result,
        })
        # removing job so that all pending requests on this job should fail.
        tasks.remove_rq_jobs()
        now = datetime.now()
        if task_result:
            filename = f'{str(now.date())}_{now.strftime("%H%M%S")}.json'
            user_id = session['google_id']
            filepath = f'{config.Config.FILE_PATH}/{user_id}/tests/{filename}'
            job_result = json.loads(task_result)
            if job_result.get('is_flight_records_from_kml'):
                del job_result['is_flight_records_from_kml']
                for filename, content in job_result.items():
                    filepath = f'{config.Config.FILE_PATH}/{user_id}/flight_records/{filename}.json'
                    _write_to_file(filepath, json.dumps(content))
                response_object.update({'is_flight_records_from_kml': True})
            else:
                job_result = task_result
                _write_to_file(filepath, job_result)
            response_object.update({'filename': filename})
        else:
            logging.info('Task result not available yet..')
    return response_object


@webapp.route('/report', methods=['POST'])
def get_report():
    user_id = session['google_id']
    output_path = f'{config.Config.FILE_PATH}/{user_id}/tests'
    try:
        output_files = os.listdir(output_path)
        output_files = [os.path.join(output_path, f) for f in output_files]
        output_files.sort(key=lambda x: os.path.getmtime(x), reverse=True)
        latest_file = output_files[0]
        with open(latest_file) as f:
            content = f.read()
            if content:
                output = make_response(content)
                output.headers['Content-Disposition'] = f'attachment; filename={os.path.basename(f.name)}'
                output.headers['Content-type'] = 'text/csv'
                return output
    except FileNotFoundError as e:
        logging.exception(e)
        return {'error': 'Error downloading file'}
    return {'Error': 'Error getting result'}


@webapp.route('/result_download/<string:filename>', methods=['POST', 'GET'])
def download_test(filename):
    user_id = session['google_id']
    filepath = f'{config.Config.FILE_PATH}/{user_id}/tests/{filename}'
    content = ''
    with open(filepath) as f:
        content = f.read()
    if content:
        output = make_response(content)
        output.headers['Content-Disposition'] = f'attachment; filename={filename}'
        output.headers['Content-type'] = 'text/csv'
        return output
    return {'error': 'Error downloading file'}

@webapp.route('/history')
def get_test_history():
    user_id = session['google_id']
    output_path = f'{config.Config.FILE_PATH}/{user_id}/tests'
    try:
        executed_tests = os.listdir(output_path)
    except:
        executed_tests = []
    return {'tests': executed_tests}


@webapp.route('/flight_records', methods=['POST'])
def upload_flight_state_files():
    """Upload files."""
    files = request.files.getlist('files[]')
    user_id = session['google_id']
    flight_records_path = f'{config.Config.FILE_PATH}/{user_id}/flight_records'
    if not os.path.isdir(flight_records_path):
        os.makedirs(flight_records_path)
    kml_files_path = f'{config.Config.FILE_PATH}/{user_id}/kml_files'
    if not os.path.isdir(kml_files_path):
        os.makedirs(kml_files_path)
    kml_files = []
    for file in files:
        if file:
            filename = secure_filename(file.filename)
            if filename.endswith('.json'):
                file_path = os.path.join(flight_records_path, filename)
                file.save(file_path)
            elif filename.endswith('.kml'):
                file_path = os.path.join(kml_files_path, filename)
                file.save(file_path)
                kml_files.append(file_path)
    logging.info(f'KML files: {kml_files}')
    if kml_files:
        return redirect(url_for('._process_kml', kml_files=json.dumps(kml_files)), code=307)
    return redirect(url_for('.tests'))

@webapp.route('/process_kml', methods=['POST'])
def _process_kml():
    kml_files = request.args['kml_files']
    user_id = session['google_id']
    flight_records_path = f'{config.Config.FILE_PATH}/{user_id}/flight_records'
    kml_jobs = []
    for file in json.loads(kml_files):
        with open(file, 'rb') as fo:
            content = fo.read()
            job_id = _process_kml_files_task(content, flight_records_path)
            kml_jobs.append(job_id)
    for job_id in kml_jobs:
        get_result(job_id)
    return redirect(url_for('.tests'))


@webapp.route('/delete', methods=['GET', 'POST'])
def delete_file():
    data = json.loads(request.get_data())
    filename = data.get('filename')
    if filename:
        user_id = session['google_id']
        file = f'{config.Config.FILE_PATH}/{user_id}/flight_records/{filename}'
        if os.path.exists(file):
            os.remove(file)
        else:
            raise 'File not found'
    return {'deleted': filename}


@webapp.route('/status')
def status():
    return 'Mock Host Service Provider ok {}'.format(
        versioning.get_code_version())


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
