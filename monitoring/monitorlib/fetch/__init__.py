import datetime
import json
from typing import Dict, Optional

import requests
import yaml
from yaml.representer import Representer

from monitoring.monitorlib import infrastructure


class RequestDescription(dict):
  @property
  def token(self) -> Dict:
    return infrastructure.get_token_claims(self.get('headers', {}))
yaml.add_representer(RequestDescription, Representer.represent_dict)


def describe_request(req: requests.PreparedRequest,
                     initiated_at: datetime.datetime) -> RequestDescription:
  headers = {k: v for k, v in req.headers.items()}
  info = {
    'method': req.method,
    'url': req.url,
    'initiated_at': initiated_at.isoformat(),
    'headers': headers,
  }
  body = req.body.decode('utf-8') if req.body else None
  try:
    if body:
      info['json'] = json.loads(body)
    else:
      info['body'] = body
  except ValueError:
    info['body'] = body
  return RequestDescription(info)


class ResponseDescription(dict):
  @property
  def status_code(self) -> int:
    return self['code']
yaml.add_representer(ResponseDescription, Representer.represent_dict)


def describe_response(resp: requests.Response) -> ResponseDescription:
  headers = {k: v for k, v in resp.headers.items()}
  info = {
    'code': resp.status_code,
    'headers': headers,
    'elapsed_s': resp.elapsed.total_seconds(),
    'reported': datetime.datetime.utcnow().isoformat(),
  }
  try:
    info['json'] = resp.json()
  except ValueError:
    info['body'] = resp.content.decode('utf-8')
  return ResponseDescription(info)


class Interaction(dict):
  @property
  def status_code(self) -> int:
    return self['response'].status_code

  @property
  def json_result(self) -> Optional[Dict]:
    return self['response'].get('json', None)

  @property
  def initiated_at(self) -> datetime.datetime:
    return self['request']['initiated_at']

  @property
  def completed_at(self) -> datetime.datetime:
    return self['response']['reported']
yaml.add_representer(Interaction, Representer.represent_dict)


def describe_interaction(resp: requests.Response,
                         initiated_at: datetime.datetime) -> Interaction:
  return Interaction({
    'request': describe_request(resp.request, initiated_at),
    'response': describe_response(resp),
  })
