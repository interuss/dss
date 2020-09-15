import datetime
import logging
import os
from typing import Dict, Tuple

import flask
from termcolor import colored
import yaml

from monitoring.monitorlib import infrastructure, versioning, geo
from monitoring.tracer import check_rid_flights, diff
from . import context, webapp


logging.basicConfig()
_logger = logging.getLogger('tracer.notifications')
_logger.setLevel(logging.DEBUG)

RESULT = ('', 204)


def _get_request_info(request: flask.Request) -> Dict:
  headers = {k: v for k, v in request.headers}
  info = {
    'method': request.method,
    'url': request.url,
    'timestamp': datetime.datetime.utcnow().isoformat(),
    'token': infrastructure.get_token_claims(headers),
    'headers': headers,
  }
  try:
    info['json'] = request.json
  except ValueError:
    info['body'] = request.data.encode('utf-8')
  return info


def _print_time_range(t0: str, t1: str) -> str:
  if not t0 and not t1:
    return ''
  now = datetime.datetime.utcnow()
  if t0.endswith('Z'):
    t0 = t0[0:-1]
  if t1.endswith('Z'):
    t1 = t1[0:-1]
  try:
    t0dt = datetime.datetime.fromisoformat(t0) - now
    t1dt = datetime.datetime.fromisoformat(t1) - now
    return ' {} to {}'.format(diff.format_timedelta(t0dt),
                              diff.format_timedelta(t1dt))
  except ValueError as e:
    return ''


@webapp.route('/v1/uss/identification_service_areas/<id>', methods=['POST'])
def rid_isa_notification(id: str) -> Tuple[str, int]:
  """Implements RID ISA notification receiver."""
  log_name = context.resources.logger.log_new('notify_isa', _get_request_info(flask.request))

  claims = infrastructure.get_token_claims({k: v for k, v in flask.request.headers})
  owner = claims.get('sub', '<No owner in token>')
  label = colored('ISA', 'cyan')
  try:
    json = flask.request.json
    if 'service_area' in json:
      isa = json['service_area']
      owner_body = isa.get('owner', None)
      if owner_body and owner_body != owner:
        owner = '{} token|{} body'.format(owner, owner_body)
      version = isa.get('version', '<Unknown version>')
      time_range = _print_time_range(isa.get('time_start', None), isa.get('time_end', None))
      _logger.info('{} {} v{} ({}) updated{} -> {}'.format(label, id, version, owner, time_range, log_name))
    else:
      _logger.info('{} {} ({}) deleted -> {}'.format(label, id, owner, log_name))
  except ValueError as e:
    _logger.error('{} {} ({}) unable to decode JSON: {} -> {}'.format(label, id, owner, e, log_name))

  return RESULT


@webapp.route('/uss/v1/operations', methods=['POST'])
def scd_operation_notification() -> Tuple[str, int]:
  """Implements SCD Operation notification receiver."""
  log_name = context.resources.logger.log_new('notify_op', _get_request_info(flask.request))

  claims = infrastructure.get_token_claims({k: v for k, v in flask.request.headers})
  owner = claims.get('sub', '<No owner in token>')
  label = colored('Operation', 'blue')
  try:
    json = flask.request.json
    id = json.get('operation_id', '<Unknown ID>')
    if 'operation' in json:
      op = json['operation']
      version = '<Unknown version>'
      ovn = '<Unknown OVN>'
      time_range = ''
      if 'reference' in op:
        op_ref = op['reference']
        owner_body = op_ref.get('owner', None)
        if owner_body and owner_body != owner:
          owner = '{} token|{} body'.format(owner, owner_body)
        version = op_ref.get('version', version)
        ovn = op_ref.get('ovn', ovn)
        time_range = _print_time_range(
          op_ref.get('time_start', {}).get('value', None),
          op_ref.get('time_end', {}).get('value', None))
      state = '<Unknown state>'
      vlos = False
      if 'details' in op:
        op_details = op['details']
        state = op_details.get('state')
        vlos = op_details.get('vlos', vlos)
      vlos_text = 'VLOS' if vlos else 'BVLOS'
      _logger.info('{} {} {} {} v{} ({}) OVN[{}] updated{} -> {}'.format(
        label, state, vlos_text, id, version, owner, ovn, time_range, log_name))
    else:
      _logger.info('{} {} ({}) deleted -> {}'.format(label, id, owner, log_name))
  except ValueError as e:
    _logger.error('{} ({}) unable to decode JSON: {} -> {}'.format(label, owner, e, log_name))

  return RESULT


@webapp.route('/uss/v1/constraints', methods=['POST'])
def scd_constraint_notification() -> Tuple[str, int]:
  """Implements SCD Constraint notification receiver."""
  log_name = context.resources.logger.log_new('notify_constraint', _get_request_info(flask.request))

  claims = infrastructure.get_token_claims({k: v for k, v in flask.request.headers})
  owner = claims.get('sub', '<No owner in token>')
  label = colored('Constraint', 'magenta')
  try:
    json = flask.request.json
    id = json.get('constraint_id', '<Unknown ID>')
    if 'constraint' in json:
      constraint = json['constraint']
      version = '<Unknown version>'
      ovn = '<Unknown OVN>'
      time_range = ''
      if 'reference' in constraint:
        constraint_ref = constraint['reference']
        owner_body = constraint_ref.get('owner', None)
        if owner_body and owner_body != owner:
          owner = '{} token|{} body'.format(owner, owner_body)
        version = constraint_ref.get('version', version)
        ovn = constraint_ref.get('ovn', ovn)
        time_range = _print_time_range(
          constraint_ref.get('time_start', {}).get('value', None),
          constraint_ref.get('time_end', {}).get('value', None))
      type = '<Unspecified type>'
      if 'details' in constraint:
        constraint_details = constraint['details']
        type = constraint_details.get('type')
      _logger.info('{} {} {} v{} ({}) OVN[{}] updated{} -> {}'.format(
        label, type, id, version, owner, ovn, time_range, log_name))
    else:
      _logger.info('{} {} ({}) deleted -> {}'.format(label, id, owner, log_name))
  except ValueError as e:
    _logger.error('{} ({}) unable to decode JSON: {} -> {}'.format(label, owner, e, log_name))

  return RESULT


@webapp.route('/status')
def status():
  return 'Ok {}'.format(versioning.get_code_version())


@webapp.route('/logs')
@webapp.route('/')
def list_logs():
  logs = list(reversed(sorted(os.listdir(context.resources.logger.log_path))))
  response = flask.make_response(flask.render_template('logs.html', logs=logs))
  response.headers['Cache-Control'] = 'no-cache, no-store, must-revalidate'
  response.headers['Pragma'] = 'no-cache'
  return response


def _redact_log(obj):
  if isinstance(obj, dict):
    result = {}
    for k, v in obj.items():
      if k.lower() == 'authorization' and isinstance(v, str):
        result[k] = '.'.join(v.split('.')[0:-1]) + '.REDACTED'
      else:
        result[k] = _redact_log(v)
    return result
  elif isinstance(obj, str):
    return obj
  elif isinstance(obj, list):
    return [_redact_log(item) for item in obj]
  else:
    return obj


@webapp.route('/logs/<log>')
def logs(log):
  logfile = os.path.join(context.resources.logger.log_path, log)
  if not os.path.exists:
    flask.abort(404)
  with open(logfile, 'r') as f:
    objs = [obj for obj in yaml.full_load_all(f)]
  if len(objs) == 1:
    obj = objs[0]
  else:
    obj = {'entries': objs}
  return flask.render_template('log.html', log=_redact_log(obj), title=logfile)


@webapp.route('/rid_poll', methods=['GET'])
def get_rid_poll():
  return flask.render_template('rid_poll.html')


@webapp.route('/rid_poll', methods=['POST'])
def request_rid_poll():
  if 'area' not in flask.request.form:
    flask.abort(400, 'Missing area')

  try:
    area = geo.make_latlng_rect(flask.request.form['area'])
  except ValueError as e:
    flask.abort(400, str(e))

  result = check_rid_flights.get_all_flights(context.resources, area, flask.request.form.get('include_recent_positions'))
  log_name = context.resources.logger.log_new('clientrequest_getflights', result)
  return flask.redirect(flask.url_for('logs', log=log_name))


@webapp.route('/favicon.ico')
def favicon():
  flask.abort(404)


@webapp.route('/<path:u_path>', methods=['GET', 'PUT', 'POST', 'DELETE'])
def catch_all(u_path) -> Tuple[str, int]:
  log_name = context.resources.logger.log_new('uss_badroute', _get_request_info(flask.request))

  claims = infrastructure.get_token_claims({k: v for k, v in flask.request.headers})
  owner = claims.get('sub', '<No owner in token>')
  label = colored('Bad route', 'red')
  _logger.error('{} to {} ({}): {}'.format(label, u_path, owner, log_name))

  return RESULT


context.init()
