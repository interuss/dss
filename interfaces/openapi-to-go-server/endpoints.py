import dataclasses
from typing import Dict, List

import data_types


def capitalize_first_letter(s: str) -> str:
  return s[0].upper() + s[1:] if s else s


@dataclasses.dataclass
class PathParameter:
  name: str
  description: str
  go_type: str

  @property
  def go_field_name(self) -> str:
    return self.name[0].upper() + self.name[1:]


@dataclasses.dataclass
class AuthorizationOption:
  required_scopes: List[str]


@dataclasses.dataclass
class Security:
  schemes: Dict[str, List[AuthorizationOption]]


@dataclasses.dataclass
class Response:
  code: int
  description: str
  json_body_type: str

  @property
  def response_set_field(self) -> str:
    return 'Response{}'.format(self.code)


@dataclasses.dataclass
class Endpoint:
  path: str
  summary: str
  description: str
  operation_id: str
  tags: List[str]
  security: Security
  verb: str
  path_parameters: List[PathParameter]
  json_request_body_type: str
  responses: List[Response]

  @property
  def handler_interface_name(self) -> str:
    if self.operation_id:
      return capitalize_first_letter(self.operation_id)
    else:
      return capitalize_first_letter(self.verb.lower()) + data_types.snake_case_to_pascal_case(self.path.replace('{', '').replace('}', '').replace('/', '_'))

  @property
  def response_type_name(self) -> str:
    return self.handler_interface_name + 'ResponseSet'

  @property
  def request_type_name(self) -> str:
    return self.handler_interface_name + 'Request'


def make_endpoints(name: str, schema: Dict, ensure_500: bool=True) -> List[Endpoint]:
  endpoints: List[Endpoint] = []

  summary = schema.get('summary', '')
  description = schema.get('description', '')

  path_parameters: List[PathParameter] = []
  for parameter in schema.get('parameters', []):
    parameter_name = parameter['name']
    parameter_description = parameter.get('description', '')
    parameter_in = parameter['in']
    if 'schema' in parameter:
      parameter_field, further_types = data_types.make_object_field('', parameter_name, parameter['schema'], set())
      if further_types:
        raise NotImplementedError('Endpoint path parameters may not currently be non-primitive')
      parameter_type = parameter_field.go_type
    else:
      parameter_type = 'string'
    if parameter_in == 'path':
      path_parameters.append(PathParameter(name=parameter_name, description=parameter_description, go_type=parameter_type))
    else:
      raise NotImplementedError('Parameter in "{}" (`{}`) not yet implemented'.format(parameter_in, parameter_name))

  for verb in ('get', 'put', 'post', 'delete'):
    if verb not in schema:
      continue
    action = schema[verb]
    verb_summary = action.get('summary', summary)
    verb_description = action.get('description', description)
    operation_id = action.get('operationId', '')
    tags = action.get('tags', [])
    request_body_type = data_types.get_data_type_name(action.get('requestBody', {}).get('content', {}).get('application/json', {}).get('schema', {}).get('$ref', ''), 'requestBody')
    security = Security(schemes={})
    for security_option in action.get('security', []):
      for scheme, scopes in security_option.items():
        options = security.schemes.get(scheme, [])
        options.append(AuthorizationOption(required_scopes=scopes))
        security.schemes[scheme] = options
    responses: List[Response] = []
    for code, response in action.get('responses', {}).items():
      responses.append(Response(
        code=int(code),
        description=response.get('description', ''),
        json_body_type=data_types.get_data_type_name(response.get('content', {}).get('application/json', {}).get('schema', {}).get('$ref', ''), 'response body')))
    if ensure_500 and 500 not in {r.code for r in responses}:
      responses.append(Response(code=500, description='Internal server error', json_body_type='InternalServerErrorBody'))

    endpoints.append(Endpoint(
      path=name,
      summary=verb_summary,
      description=verb_description,
      operation_id=operation_id,
      tags=tags,
      security=security,
      verb=verb,
      path_parameters=path_parameters,
      json_request_body_type=request_body_type,
      responses=responses
    ))

  return endpoints
