import flask
import functools
import json
import logging
import os
import pathlib
import requests
import uuid

from . import config, resources, tasks
from . import forms, messages

from datetime import datetime
from google.oauth2 import id_token
from google_auth_oauthlib.flow import Flow
import google.auth.transport.requests
from flask import (
    render_template,
    request,
    make_response,
    redirect,
    url_for,
    session,
    abort,
    flash,
)
from functools import wraps
from pip._vendor import cachecontrol
from werkzeug.exceptions import HTTPException
from werkzeug.utils import secure_filename
from google.auth.transport import requests as auth_requests

from monitoring.monitorlib import versioning, auth_validation
from monitoring.uss_qualifier.webapp import webapp


_SERVICE_ACCOUNT_TOKEN_HEADER = 'access_token'

client_secrets_file = os.path.join(
    pathlib.Path(__file__).parent, "client_secret.json")


def _create_oauth_flow(state=None):
    flow = Flow.from_client_secrets_file(
        client_secrets_file=client_secrets_file,
        scopes=[
            "https://www.googleapis.com/auth/userinfo.profile",
            "https://www.googleapis.com/auth/userinfo.email",
            "openid",
        ],
        state=state)
    flow.redirect_uri = f"{webapp.config.get(config.KEY_USS_QUALIFIER_HOST_URL)}/login_callback"
    return flow


def is_service_account():
    """Returns True if the logged in user is a service account."""
    return flask.session.get('is_service_account', False)


def login_required(fn=None, *, origin_path=None):
    """Decorator that ensures the user is authenticated with a browser session.

    Can be used with or without arguments.

    Args:
      fn: function to wrap.
      oneof: a list of backend.models.Permissions. The logged-in user must have at
        least one of these permissions. If None, no permission check is done.

    Returns:
      Wrapped function.
    """

    def _outer(fn):
        """Decorator."""

        @functools.wraps(fn)
        def _inner(*args, **kwargs):
            """Decorator."""

            user_email = None

            if credentials_from_session() is not None:
                logging.debug('User email from session cookie: %s',
                              flask.session['email'])
                return fn(*args, **kwargs)
            else:
                access_token = flask.request.headers.get('access_token')
                if access_token:
                    return _handle_service_account_auth(access_token, origin_path)

            logging.info('User is not authenticated - starting OAuth flow')
            flow = _create_oauth_flow(state=origin_path)
            authorization_url, state = flow.authorization_url(
                login_hint=user_email)
            flask.session['state'] = state or origin_path
            return flask.redirect(authorization_url)

        return _inner

    if fn is not None:
        return _outer(fn)
    else:
        return _outer


def _handle_service_account_auth(token, origin_path=None):
    """Login handler for service accounts."""

    if _SERVICE_ACCOUNT_TOKEN_HEADER not in flask.request.headers:
        return 'Credential error', 403  # Deliberately vauge error message.

    creds = google.oauth2.credentials.Credentials(token)
    sess = google.auth.transport.requests.AuthorizedSession(creds)
    resp = sess.get(
        f'https://oauth2.googleapis.com/tokeninfo?access_token={token}')
    try:
        # TODO: Still need to handle in a better way
        resp.raise_for_status()
    except requests.exceptions.HTTPError as e:
        return e, 403

    # Populate the session cookie.
    userinfo = resp.json()
    flask.session.clear()
    flask.session.permanent = True
    flask.session['user_id'] = userinfo['sub']
    flask.session['email'] = userinfo['email']
    flask.session['access_token'] = token
    flask.session['is_service_account'] = True

    logging.info('Successful sign in for service account %s',
                 userinfo['email'])

    return flask.redirect(origin_path or "/")


@webapp.route("/login_callback")
def login_callback():
    """Handles the oauth2 callback."""

    flow = _create_oauth_flow(state=flask.session['state'])
    flow.fetch_token(authorization_response=flask.request.url)

    if not flow.credentials:
        logging.warning('OAuth callback without valid credentials')
        return 'No credentials', 403

    id_info = google.oauth2.id_token.verify_oauth2_token(
        flow.credentials.id_token, google.auth.transport.requests.Request(),
        flow.client_config['client_id'])

    # Store the important bits in the signed session cookie.
    flask.session.clear()  # Remove the CSRF token as it's not needed any more.
    flask.session.permanent = True
    flask.session['user_id'] = id_info['sub']
    flask.session['user_id'] = id_info['sub']
    flask.session['email'] = id_info['email']
    flask.session['name'] = id_info['name']
    flask.session['access_token'] = flow.credentials.token

    logging.info('Successful sign in for email %s', id_info['email'])

    return redirect(request.args.get('next') or "/")


def credentials_from_session():
    """Returns the user's credentials from the session cookie."""
    if 'access_token' in flask.session and flask.session['access_token']:
        return google.oauth2.credentials.Credentials(flask.session['access_token'])
    return None


@webapp.route("/logout")
def logout():
    """Revokes the access_token, clears the session cookie and redirects to /."""

    logging.info('Logging out user %s', flask.session.get('email'))

    if 'access_token' in flask.session and flask.session['access_token']:
        flask.session.clear()
    return render_template("logout.html")


@webapp.before_request
def _process_completed_bg_jobs():
    _reload_latest_kmls_from_redis()
    _reload_latest_test_run_outcomes_from_redis()


def _initialize_background_test_runs(
    user_config, auth_spec, input_files_content, input_files, user_id, debug
):
    now = datetime.now()
    testruns_id = f'{str(now.date())}_{now.strftime("%H%M%S%f")}.json'
    job = resources.qualifier_queue.enqueue(
        "monitoring.uss_qualifier.webapp.tasks.call_test_executor",
        user_config,
        auth_spec,
        input_files_content,
        testruns_id,
        debug,
        config.Config.SCD_TEST_DEFINITIONS_FILE_PATH,
    )
    task_id = job.get_id()
    task = tasks.get_rq_job(task_id)
    task_details = {
        "test_run_id": testruns_id,
        "specifications": {
            "flight_records": input_files,
            "auth_spec": auth_spec,
            "user_config": json.loads(user_config),
        },
        "task": {"id": task_id, "status": task.get_status()},
        "user_id": user_id,
        "status_message": "A task has been started in the background",
    }
    resources.redis_conn.hset(
        resources.REDIS_KEY_TEMP_LOGS, testruns_id, json.dumps(task_details)
    )
    return task_details


def _get_running_jobs():
    registry = resources.qualifier_queue.started_job_registry
    running_job = registry.get_job_ids()
    if running_job:
        return running_job[0]


def _process_kml_files_task(kml_file, output_path):
    with open(kml_file, "rb") as fo:
        kml_content = fo.read()
        job = resources.qualifier_queue.enqueue(
            "monitoring.uss_qualifier.webapp.tasks.call_kml_processor",
            kml_content,
            output_path,
        )
        job_id = job.get_id()
        # Adding an empty job id in the queue to mark a latest kml job.
        resources.redis_conn.hset(
            resources.REDIS_KEY_UPLOADED_KMLS, job_id, "")
        return job_id


def _get_user_local_config():
    """Get user's last saved specs."""
    # TODO: replace hardcoded user_id with session user.
    user_id = session['user_id']
    user_config_file = (
        f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/user_config.json"
    )
    auth_spec = ""
    config_spec = ""
    if os.path.isfile(user_config_file):
        with open(user_config_file) as fo:
            file_content = json.loads(fo.read())
            auth_spec = file_content["auth"]
            config_spec = file_content["config"]
    return auth_spec, config_spec


def _update_user_local_config(auth_spec, config_spec):
    """Saves user's local config in  profile specific folder."""
    # TODO: replace hardcoded user_id with session user.
    user_id = session['user_id']
    user_config_file = (
        f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/user_config.json"
    )
    user_config = {"auth": auth_spec, "config": config_spec}
    _write_to_file(user_config_file, json.dumps(user_config))


@webapp.route("/info")
def info():
    return render_template("info.html")


@webapp.route("/", methods=["GET"])
@login_required
def tests():
    files = []
    user_id = session['user_id']
    running_job = _get_running_jobs()
    flight_record_data = get_flight_records()

    if flight_record_data.get("flight_records"):
        files = [(x, x) for x in flight_record_data["flight_records"]]

    auth_spec, config_spec = _get_user_local_config()
    form = forms.TestsExecuteForm()
    form.flight_records.choices = files
    form.auth_spec.data = auth_spec
    form.user_config.data = config_spec
    test_run_history = _get_test_runs_logs(user_id)
    completed_tests = [
        t["test_run_id"] for t in test_run_history if t["task"]["status"] == "finished"
    ]
    data = {"tests": completed_tests}
    if running_job:
        data.update({"job_id": running_job})
    return render_template("tests.html", title="Execute tests", form=form, data=data)


@webapp.route("/", methods=["POST"])
@login_required
def tests_submit():
    form = forms.TestRunsForm(request.form)
    user_id = session['user_id']
    form_validation_status = _validate_config_form(form, user_id)
    if form_validation_status["err_message"]:
        return render_template(
            "tests.html",
            title="Execute tests",
            form=form,
            data=form_validation_status["details"],
        )
    else:
        run_tests(form_validation_status=form_validation_status)
        return redirect(url_for(".tests"))


def _validate_config_form(form, user_id):
    err_message = ""
    form_validation_status = {"err_message": "", "details": {}}
    if not form.validate():
        form_validation_status["err_message"] = "validation error"
        form_validation_status["details"].update(form.errors)
    else:
        flight_records_files = []
        if request.form.get("flight_records"):
            flight_records_files = [
                i.strip() for i in (request.form["flight_records"]).split(",")
            ]

        file_objs = []
        input_files = []
        for record in flight_records_files:
            filename = secure_filename(record)
            input_files.append(filename)
            filepath = f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/flight_records/{filename}"
            if os.path.exists(filepath):
                with open(filepath) as fo:
                    file_objs.append(fo.read())
            else:
                existing_flights = get_flight_records()
                if existing_flights and existing_flights["flight_records"]:
                    err_message = "%s\n%s" % (
                        messages.FLIGHT_RECORD_NOT_FOUND.format(
                            filename=filename),
                        messages.FLIGHT_RECORDS_EXISTING.format(
                            flight_records="\n".join(
                                existing_flights["flight_records"])
                        ),
                    )
                else:
                    err_message = messages.FLIGHT_RECORD_NOT_FOUND.format(
                        filename=filename
                    )
        form_validation_status["err_message"] = err_message
        form_validation_status["file_objs"] = file_objs
        form_validation_status["input_files"] = input_files
    return form_validation_status


@webapp.route("/api/test_runs", methods=["POST"])
@login_required(origin_path='/api/test_runs')
def run_tests(form_validation_status=None):
    running_job = _get_running_jobs()
    user_id = session['user_id']
    if running_job:
        return {
            "task": {
                "id": running_job,
            },
            "status_message": "A job already running in the background, report is not ready yet",
            "user_id": user_id,
            "report": None,
        }
    # TODO:(pratibha) user_id hardcoded until Auth is fixed.
    form = forms.TestRunsForm(request.form)
    if not form_validation_status:
        form_validation_status = _validate_config_form(form, user_id)
        if form_validation_status["err_message"]:
            forms.json_abort(
                400,
                form_validation_status["err_message"],
                details=form_validation_status["details"],
            )
    auth_spec = form.auth_spec.data
    user_config = form.user_config.data
    _update_user_local_config(auth_spec, user_config)
    task_details = _initialize_background_test_runs(
        user_config,
        auth_spec,
        form_validation_status["file_objs"],
        form_validation_status["input_files"],
        user_id,
        debug=form.sample_report.data,
    )
    return {"test_run": task_details}


def _get_test_runs_logs(user_id):
    tests_logs = resources.redis_conn.hgetall(
        f"{user_id}-{resources.REDIS_KEY_TEST_RUN_LOGS}"
    )
    tests_logs = resources.decode_redis(tests_logs)
    logs = []
    for _, log in tests_logs.items():
        test_run_logs = json.loads(log)
        task_id = test_run_logs["task"]["id"]
        task = tasks.get_rq_job(task_id)
        if task:
            task_status = task.get_status()
            test_run_logs["task"]["status"] = task_status
            task_result = json.loads(task.result) if task.result else None
            test_run_logs["report"] = task_result
        else:
            test_run_logs["task"]["status"] = "finished"
        logs.append(test_run_logs)
    return logs


@webapp.route("/api/test_runs", methods=["GET"])
@login_required(origin_path='/api/test_runs')
def get_tests_history():
    user_id = session['user_id']
    test_runs_logs = _get_test_runs_logs(user_id)
    return {"test_runs": test_runs_logs}


@webapp.route("/api/test_runs/<string:test_id>", methods=["GET"])
@login_required(origin_path='/api/test_runs')
def get_test_runs_details(test_id):
    user_id = session['user_id']
    test_runs_logs = _get_test_runs_logs(user_id)
    result_set = list(
        filter(lambda p: p["test_run_id"] == test_id, test_runs_logs))
    if result_set:
        return {"test_run": result_set[0]}
    abort(400, f"test_run_id: {test_id} does not exist.")


@webapp.route("/api/tasks/<string:task_id>", methods=["GET"])
def get_task_status(task_id):
    # If task ID is for Imported KML job, the status is maintained through following redis queue.
    task_list = resources.redis_conn.hgetall(resources.REDIS_KEY_UPLOADED_KMLS)
    curr_task_id = task_id.encode("utf-8")
    if task_list.get(curr_task_id):
        return {"task": json.loads(task_list[curr_task_id])}
    else:
        if session.get("completed_job") == task_id:
            abort(400, "Request already processed")
        task = tasks.get_rq_job(task_id)
        response_object = {"task": {}}
        if task:
            response_object["task"] = {
                "task_id": task.get_id(),
                "task_status": task.get_status(),
                "task_result": task.result,
            }
        else:
            abort(400, f"task_id: {task_id} does not exist.")
        if task.get_status() == "finished":
            session["completed_job"] = task_id
            # removing job so that all the pending requests on this job should abort.
            tasks.remove_rq_job(task_id)
        return response_object


def _reload_latest_kmls_from_redis():
    latest_kmls = resources.redis_conn.hgetall(
        resources.REDIS_KEY_UPLOADED_KMLS)
    user_id = session.get('user_id', 'localuser')
    if latest_kmls:
        task_status = {}
        for task_id, val in latest_kmls.items():
            if not is_kml_bg_task_processed(val):
                generated_flight_records = []
                task_details = _get_task_status(task_id)
                if task_details:
                    if task_details["task_status"] == "finished":
                        flight_data = json.loads(task_details["task_result"])
                        if not isinstance(flight_data, dict):
                            flight_data = json.loads(flight_data)
                        if flight_data.get("is_flight_records_from_kml"):
                            del flight_data["is_flight_records_from_kml"]
                            for filename, content in flight_data.items():
                                filepath = f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/flight_records"
                                json_file_version = _get_latest_file_version(
                                    filepath, filename
                                )
                                _write_to_file(json_file_version,
                                               json.dumps(content))
                                _, generated_flight_file = os.path.split(
                                    json_file_version
                                )
                                generated_flight_records.append(
                                    generated_flight_file)
                        task_status[
                            "generated_flight_records"
                        ] = generated_flight_records
                        task_status["task_status"] = task_details["task_status"]
                        task_status["task_id"] = task_id.decode("utf-8")
                        task_status["processed"] = True
                        resources.redis_conn.hset(
                            resources.REDIS_KEY_UPLOADED_KMLS,
                            task_id.decode("utf-8"),
                            json.dumps(task_status),
                        )


def is_kml_bg_task_processed(task_result):
    if not task_result:
        return False
    if not isinstance(task_result, dict):
        task_result = json.loads(task_result)
        # In case task result is ready, but not processed yet, result is in the form of binary wrapped in string format.
        if not isinstance(task_result, dict):
            task_result = json.loads(task_result)
    if task_result.get("processed"):
        return True
    if task_result.get("task_status") and task_result["task_status"] != "failed":
        return False
    return True


def _reload_latest_test_run_outcomes_from_redis():
    latest_test_runs_report = resources.redis_conn.hgetall(
        resources.REDIS_KEY_TEST_RUNS
    )
    user_id = session.get('user_id', 'localuser')
    temp_logs = resources.redis_conn.hgetall(resources.REDIS_KEY_TEMP_LOGS)
    temp_logs = resources.decode_redis(temp_logs)

    latest_test_runs_report = resources.decode_redis(latest_test_runs_report)
    for filename in temp_logs:
        filepath = (
            f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/tests/{filename}"
        )

        temp_logs[filename] = json.loads(temp_logs[filename])

        if filename in latest_test_runs_report:
            test_result = latest_test_runs_report[filename]
            if isinstance(test_result, bytes):
                test_result = test_result.decode("utf-8")
            _write_to_file(filepath, test_result)
            temp_logs[filename].update(
                {
                    "report": json.loads(test_result),
                    "test_run_id": filename,
                    "status_message": "Report Ready",
                }
            )
        resources.redis_conn.hset(
            f"{user_id}-{resources.REDIS_KEY_TEST_RUN_LOGS}",
            filename,
            json.dumps(temp_logs[filename]),
        )
        resources.redis_conn.delete(resources.REDIS_KEY_TEST_RUN_LOGS)
        resources.redis_conn.delete(resources.REDIS_KEY_TEMP_LOGS)


def _get_latest_file_version(filepath, filename):
    if not filename.endswith(".json"):
        curr_file_path = f"{filepath}/{filename}.json"
    else:
        curr_file_path = f"{filepath}/{filename}"
    if os.path.exists(curr_file_path):
        version_counter = 0
        while True:
            version_counter += 1
            filename = filename.replace(".json", "")
            curr_file_path = f"{filepath}/{filename}_{str(version_counter)}.json"
            if not os.path.exists(curr_file_path):
                break
    return curr_file_path


def get_flight_records():
    data = {"flight_records": [], "message": ""}
    user_id = session["user_id"]
    folder_path = f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/flight_records"
    if not os.path.isdir(folder_path):
        data["message"] = "Flight records not available."
    else:
        flight_records = [f for f in os.listdir(
            folder_path) if f.endswith(".json")]
        data["flight_records"] = flight_records
    return data


def _write_to_file(filepath, content):
    os.makedirs(os.path.dirname(filepath), exist_ok=True)
    with open(filepath, "w") as f:
        f.write(content)


def _get_task_status(task_id):
    if isinstance(task_id, bytes):
        task_id = task_id.decode("utf-8")
    task = tasks.get_rq_job(task_id)
    task_details = {}
    if task:
        task_details = {
            "task_id": task.get_id(),
            "task_status": task.get_status(),
            "task_result": task.result,
        }
    return task_details


@webapp.route("/result/<string:job_id>", methods=["GET", "POST"])
def get_result(job_id):
    if session.get("completed_job") == job_id:
        abort(400, "Request already processed")
    response_object = _get_task_status(job_id)
    if response_object and response_object["task_status"] == "finished":
        session["completed_job"] = job_id
        task_result = response_object["task_result"]
        # removing job so that all the pending requests on this job should abort.
        tasks.remove_rq_job(job_id)
        now = datetime.now()
        if task_result:
            user_id = session["user_id"]
            job_result = json.loads(task_result)
            if job_result.get("is_flight_records_from_kml"):
                del job_result["is_flight_records_from_kml"]
                for filename, content in job_result.items():
                    filepath = f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/flight_records"
                    json_file_version = _get_latest_file_version(
                        filepath, filename)
                    _write_to_file(json_file_version, json.dumps(content))
                response_object.update({"is_flight_records_from_kml": True})
            else:
                filename = f'{str(now.date())}_{now.strftime("%H%M%S")}.json'
                filepath = f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/tests/{filename}"
                job_result = task_result
                _write_to_file(filepath, job_result)
            response_object.update({"filename": filename})
        else:
            logging.info("Task result not available yet..")
    return response_object


@webapp.route("/report", methods=["POST"])
def get_report():
    user_id = session["user_id"]
    output_path = f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/tests"
    try:
        output_files = os.listdir(output_path)
        output_files = [os.path.join(output_path, f) for f in output_files]
        output_files.sort(key=lambda x: os.path.getmtime(x), reverse=True)
        latest_file = output_files[0]
        with open(latest_file) as f:
            content = f.read()
            if content:
                output = make_response(content)
                output.headers[
                    "Content-Disposition"
                ] = f"attachment; filename={os.path.basename(f.name)}"
                output.headers["Content-type"] = "text/csv"
                return output
    except FileNotFoundError as e:
        logging.exception(e)
        return {"error": "Error downloading file"}
    return {"Error": "Error getting result"}


@webapp.route("/result_download/<string:filename>", methods=["POST", "GET"])
def download_test(filename):
    user_id = session["user_id"]
    filepath = f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/tests/{filename}"
    content = ""
    with open(filepath) as f:
        content = f.read()
    if content:
        output = make_response(content)
        output.headers["Content-Disposition"] = f"attachment; filename={filename}"
        output.headers["Content-type"] = "text/csv"
        return output
    return {"error": "Error downloading file"}


@webapp.route("/upload_kmls", methods=["POST"])
def upload_kmls():
    try:
        upload_kml_flight_records()
    except HTTPException as e:
        flash(str(e), "error")
    return redirect(url_for(".tests"))


@webapp.route("/api/kml_import_jobs", methods=["POST"])
@login_required(origin_path="/api/kml_import_jobs")
def upload_kml_flight_records():
    files = request.files.getlist("files") or request.files.getlist("files[]")
    if not files:
        abort(400, "KML files not provided.")
    user_id = session['user_id']
    flight_records_path = (
        f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/flight_records"
    )
    if not os.path.isdir(flight_records_path):
        os.makedirs(flight_records_path)
    kml_files_path = f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/kml_files"
    if not os.path.isdir(kml_files_path):
        os.makedirs(kml_files_path)
    kml_files = []
    response = {}
    message = "Invalid file type"
    for file in files:
        if file:
            filename = secure_filename(file.filename)
            if filename.endswith(".kml"):
                file_path = os.path.join(kml_files_path, filename)
                file.save(file_path)
                kml_files.append(file_path)
                message = "OK"
            else:
                message = f"Invalid file extension: {filename}"
                abort(400, message)
    if kml_files:
        bg_tasks = []
        for kml_file in kml_files:
            task_id = _process_kml_files_task(kml_file, flight_records_path)
            bg_tasks.append(task_id)
        response["kml_import_job_id"] = str(uuid.uuid4())
        response["background_tasks"] = bg_tasks
    if message == "OK":
        response[
            "status_message"
        ] = "Background tasks have started to process the KML files."
    else:
        abort(400, message)
    response["kml_imported_files"] = kml_files
    return response


@webapp.route("/upload_flight_records", methods=["POST"])
def upload_flights_records():
    try:
        _ = upload_json_flight_records()
    except HTTPException as e:
        flash(str(e), "error")
    return redirect(url_for(".tests"))


@webapp.route("/api/flight_records", methods=["POST"])
@login_required(origin_path="/api/flight_records")
def upload_json_flight_records():
    files = request.files.getlist("files") or request.files.getlist("files[]")
    if not files:
        abort(400, "Flight records not provided.")
    # TODO:
    user_id = session['user_id']
    flight_records_path = (
        f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/flight_records"
    )
    if not os.path.isdir(flight_records_path):
        os.makedirs(flight_records_path)

    response = {}
    uploaded_files = []
    for file in files:
        if file:
            filename = secure_filename(file.filename)
            if filename.endswith(".json"):
                json_file_path = _get_latest_file_version(
                    flight_records_path, filename)
                file.save(json_file_path)
                uploaded_files.append(json_file_path)
            else:
                abort(400, f"Invalid file extension: {filename}")
    response["flight_record_ids"] = uploaded_files
    return response


@webapp.route("/delete", methods=["GET", "POST"])
def delete_file():
    data = json.loads(request.get_data())
    filename = data.get("filename")
    if filename:
        user_id = session["user_id"]
        file = f"{webapp.config.get(config.KEY_FILE_PATH)}/{user_id}/flight_records/{filename}"
        if os.path.exists(file):
            os.remove(file)
        else:
            raise "File not found"
    return {"deleted": filename}


@webapp.route("/status")
def status():
    return "Mock Host Service Provider ok {}".format(versioning.get_code_version())


@webapp.errorhandler(Exception)
def handle_exception(e):
    if isinstance(e, HTTPException):
        return e
    elif isinstance(e, auth_validation.InvalidScopeError):
        return (
            flask.jsonify(
                {
                    "message": "Invalid scope; expected one of {%s}, but received only {%s}"
                    % (" ".join(e.permitted_scopes), " ".join(e.provided_scopes))
                }
            ),
            403,
        )
    elif isinstance(e, auth_validation.InvalidAccessTokenError):
        return flask.jsonify({"message": e.message}), 401
    elif isinstance(e, auth_validation.ConfigurationError):
        return flask.jsonify({"message": e.message}), 500
    elif isinstance(e, ValueError):
        return flask.jsonify({"message": str(e)}), 400

    return flask.jsonify({"message": str(e)}), 500
