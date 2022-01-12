import re
from typing import List, Optional

import data_types
import endpoints


def comment(lines: List[str]) -> List[str]:
  return ['// ' + line for line in lines]


def indent(lines: List[str], level: int) -> List[str]:
  if level == 0:
    return lines
  else:
    return ['  '*level + line if line else '' for line in lines]


def header(package: str) -> List[str]:
  lines: List[str] = []
  lines.extend(comment(['This file is auto-generated; do not change as any changes will be overwritten']))
  lines.append('package {}'.format(package))
  lines.append('')
  return lines


def data_type(d_type: data_types.DataType) -> List[str]:
  lines = comment(d_type.description.split('\n')) if d_type.description else []

  if d_type.is_primitive():
    lines.append('type {} {}'.format(d_type.name, d_type.go_type))
  elif d_type.go_type == 'struct':
    lines.append('type %s struct {' % d_type.name)
    for field in d_type.fields:
      lines.extend(indent(object_field(field), 1))
      lines.append('')
    if d_type.fields:
      lines.pop()
      lines.append('}')
    else:
      lines[-1] += '}'
  elif d_type.go_type.startswith('array:'):
    lines.append('type {} []{}'.format(d_type.name, d_type.go_type[len('array:'):]))
  else:
    lines.append('type {} {}'.format(d_type.name, d_type.go_type))

  if d_type.enum_values:
    lines.append('const (')
    lines.extend(indent(['{type}_{value} {type} = "{value}"'.format(type=d_type.name, value=v) for v in d_type.enum_values], 1))
    lines.append(')')

  return lines


def object_field(field: data_types.ObjectField) -> List[str]:
  lines = comment(field.description.split('\n')) if field.description else []
  lines.append('{} {}{} `json:"{}"`'.format(field.go_name, '*' if not field.required else '', field.go_type, field.api_name))
  return lines


def implementation_interface(declared_endpoints: List[endpoints.Endpoint]) -> List[str]:
  lines: List[str] = []

  lines.append('type EmptyResponseBody struct {}')
  lines.append('')
  lines.append('type InternalServerErrorBody struct {')
  lines.extend(indent(['ErrorMessage string `json:"error_message"`'], 1))
  lines.append('}')
  lines.append('')

  # Provide security constants
  lines.extend(comment(['Security requirements']))
  lines.append('type AuthorizationOption struct {')
  lines.extend(indent(['RequiredScopes []string'], 1))
  lines.append('}')
  lines.append('type SecurityScheme []AuthorizationOption')
  lines.append('var (')

  var_body: List[str] = []
  for endpoint in declared_endpoints:
    var_body.append('%sSecurity = map[string]SecurityScheme{' % endpoint.handler_interface_name)

    init_body: List[str] = []
    for scheme, options in endpoint.security.schemes.items():
      init_body.append('"%s": []AuthorizationOption{' % scheme)
      init_body.extend(indent(['AuthorizationOption{RequiredScopes: []string{%s}},' % ', '.join('"{}"'.format(scope) for scope in option.required_scopes) for option in options], 1))
      init_body.append('},')
    var_body.extend(indent(init_body, 1))

    var_body.append('}')
  lines.extend(indent(var_body, 1))

  lines.append(')')

  # Declare request & response types for all endpoints
  for endpoint in declared_endpoints:
    lines.append('')

    # Declare request type for query
    lines.append('type {} struct {{'.format(endpoint.request_type_name))

    body: List[str] = []
    for p in endpoint.path_parameters:
      if p.description:
        body.extend(comment(p.description.split('\n')))
      body.append('{} {}'.format(p.go_field_name, p.go_type))
      body.append('')
    if endpoint.json_request_body_type:
      body.extend(comment(['The data contained in the body of this request, if it parsed correctly']))
      body.append('Body *{}'.format(endpoint.json_request_body_type))
      body.append('')
      body.extend(comment(['The error encountered when attempting to parse the body of this request']))
      body.append('BodyParseError error')
      body.append('')
    body.extend(comment(['The result of attempting to authorize this request']))
    body.append('Auth AuthorizationResult')
    lines.extend(indent(body, 1))

    lines.append('}')

    # Declare response type for query
    lines.append('type {} struct {{'.format(endpoint.response_type_name))

    for response in endpoint.responses:
      if response.description:
        lines.extend(indent(comment(response.description.split('\n')), 1))
      body_type = response.json_body_type if response.json_body_type else 'EmptyResponseBody'
      lines.extend(indent(['{} *{}'.format(response.response_set_field, body_type)], 1))
      lines.append('')
    lines.pop()

    lines.append('}')

  lines.append('')
  lines.append('type Implementation interface {')

  body: List[str] = []
  for endpoint in declared_endpoints:
    comments: List[str] = []
    if endpoint.summary and endpoint.summary != endpoint.description:
      comments.extend(endpoint.summary.split('\n'))
    if endpoint.summary and endpoint.description and endpoint.summary != endpoint.description:
      comments.append('---')
    if endpoint.description:
      comments.extend(endpoint.description.split('\n'))
    body.extend(comment(comments))

    body.append('{}(req *{}) {}'.format(endpoint.handler_interface_name, endpoint.request_type_name, endpoint.response_type_name))
    body.append('')
  body.pop()
  lines.extend(indent(body, 1))

  lines.append('}')
  return lines


def routes(declared_endpoints: List[endpoints.Endpoint], path_prefix: Optional[str]=None) -> List[str]:
  if path_prefix is None:
    path_prefix = ''
  lines: List[str] = []

  for endpoint in declared_endpoints:
    lines.append('func (s *APIRouter) {}(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {{'.format(endpoint.handler_interface_name))

    body: List[str] = []

    body.append('var req {}'.format(endpoint.request_type_name))
    body.append('')

    body.extend(comment(['Authorize request']))
    body.append('req.Auth = s.Authorizer.Authorize(w, r, &{}Security)'.format(endpoint.handler_interface_name))
    body.append('')

    if endpoint.path_parameters:
      body.extend(comment(['Parse path parameters']))
      body.append('pathMatch := exp.FindStringSubmatch(r.URL.Path)')
      for i, p in enumerate(endpoint.path_parameters):
        if p.go_type == 'string':
          body.append('req.{} = pathMatch[{}]'.format(p.go_field_name, i + 1))
        else:
          body.append('req.{} = {}(pathMatch[{}])'.format(p.go_field_name, p.go_type, i + 1))
      body.append('')

    if endpoint.json_request_body_type:
      body.extend(comment(['Parse request body']))
      body.append('req.Body = new({})'.format(endpoint.json_request_body_type))
      body.append('defer r.Body.Close()')
      body.append('req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)')
      body.append('')

    body.extend(comment(['Call implementation']))
    body.append('response := s.Implementation.{}(&req)'.format(endpoint.handler_interface_name))
    body.append('')
    body.extend(comment(['Write response to client']))
    responses = [r for r in endpoint.responses]
    for response in responses:
      body.append('if response.{} != nil {{'.format(response.response_set_field))
      body.extend(indent(['writeJson(w, {}, response.{})'.format(response.code, response.response_set_field)], 1))
      body.extend(indent(['return'], 1))
      body.append('}')
    body.append('writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})')

    lines.extend(indent(body, 1))

    lines.append('}')
    lines.append('')
  lines.pop()

  lines.append('')

  lines.append('func MakeAPIRouter(impl Implementation, auth Authorizer) APIRouter {')

  body: List[str] = []
  body.append('router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*Route, %s)}' % len(declared_endpoints))
  body.append('')
  first_assignment = True
  for i, endpoint in enumerate(declared_endpoints):
    path_regex = path_prefix + re.sub(r'{([^}]*)}', r'(?P<\1>[^/]*)', endpoint.path)
    body.append('pattern {}= regexp.MustCompile("^{}$")'.format(':' if first_assignment else '', path_regex))
    body.append('router.Routes[%d] = &Route{Pattern: pattern, Handler: router.%s}' % (i, endpoint.handler_interface_name))
    body.append('')
    first_assignment = False
  body.append('return router')
  lines.extend(indent(body, 1))

  lines.append('}')

  return lines


def example(declared_endpoints: List[endpoints.Endpoint]) -> List[str]:
  lines: List[str] = []
  lines.append('type ExampleImplementation struct {}')
  lines.append('')
  for endpoint in declared_endpoints:
    lines.append('func (*ExampleImplementation) {}(req *{}) {} {{'.format(endpoint.handler_interface_name, endpoint.request_type_name, endpoint.response_type_name))

    body: List[str] = []
    body.append('response := %s{}' % (endpoint.response_type_name))
    #body.append('response.Response500 = &InternalServerErrorBody{ErrorMessage: "Not yet implemented"}')
    body.append('response.%s = &%s{}' % (endpoint.responses[0].response_set_field, endpoint.responses[0].json_body_type))
    body.append('return response')
    lines.extend(indent(body, 1))

    lines.append('}')
    lines.append('')

  lines.append('func main() {')

  body: List[str] = []
  body.append('router1 := MakeAPIRouter(&ExampleImplementation{}, &PermissiveAuthorizer{})')
  body.append('multiRouter := MultiRouter{Routers: []*APIRouter{&router1}}')
  body.append('s := &http.Server{')

  args: List[str] = []
  args.append('Addr:           ":8080",')
  args.append('Handler:        &multiRouter,')
  args.append('ReadTimeout:    10 * time.Second,')
  args.append('WriteTimeout:   10 * time.Second,')
  args.append('MaxHeaderBytes: 1 << 20,')
  body.extend(indent(args, 1))

  body.append('}')
  body.append('log.Fatal(s.ListenAndServe())')

  lines.extend(indent(body, 1))

  lines.append('}')

  return lines
