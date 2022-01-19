import re
from typing import List, Optional

import data_types
import operations


def comment(lines: List[str]) -> List[str]:
    """Prepend comment characters to each line.

    :param lines: Lines of text to be commented
    :return: Same lines of text provided after each line is commented
    """
    return ['// ' + line for line in lines]


def indent(lines: List[str], level: int) -> List[str]:
    """Indent each line.

    :param lines: Lines of text to be indented
    :param level: Level of indent (each indent level is two spaces)
    :return: Same lines of text provided after each line is indented
    """
    if level == 0:
        return lines
    else:
        return ['  ' * level + line if line else '' for line in lines]


def header(package: str) -> List[str]:
    """Generate a Go file header for a file in the specified package.

    :param package: Name of package in which the Go file is located
    :return: Lines of text constituting the requested Go file header
    """
    lines: List[str] = []
    lines.extend(comment(['This file is auto-generated; do not change as any changes will be overwritten']))
    lines.append('package {}'.format(package))
    lines.append('')
    return lines


def data_type(d_type: data_types.DataType) -> List[str]:
    """Generate Go code defining the provided data type.

    :param d_type: Parsed API data type to render into Go code
    :return: Lines of Go code defining the provided data type
    """
    lines = comment(
        d_type.description.split('\n')) if d_type.description else []

    if d_type.is_primitive():
        lines.append('type {} {}'.format(d_type.name, d_type.go_type))
    elif d_type.go_type == 'struct':
        lines.append('type %s struct {' % d_type.name)
        for field in d_type.fields:
            lines.extend(indent(_object_field(field), 1))
            lines.append('')
        if d_type.fields:
            lines.pop()
            lines.append('}')
        else:
            lines[-1] += '}'
    else:
        lines.append('type {} {}'.format(d_type.name, d_type.go_type))

    if d_type.enum_values:
        lines.append('const (')
        lines.extend(indent(['{type}_{value} {type} = "{value}"'.format(
            type=d_type.name, value=v) for v in d_type.enum_values], 1))
        lines.append(')')

    return lines


def _object_field(field: data_types.ObjectField) -> List[str]:
    """Generate an unindented definition of the provided field in Go code.

    :param field: Data type field to render into Go code
    :return: Lines of Go code defining the provided field
    """
    lines = comment(field.description.split('\n')) if field.description else []
    lines.append('{} {}{} `json:"{}"`'.format(field.go_name,
                                              '*' if not field.required else '',
                                              field.go_type, field.api_name))
    return lines


def implementation_interface(declared_operations: List[operations.Operation]) -> List[str]:
    """Generate Go code defining the interface an API implementation must implement.

    :param declared_operations: Operations to be included in the interface
    :return: Lines of Go code defining the interface
    """
    lines: List[str] = []

    # Provide security constants
    lines.append('var (')

    var_body: List[str] = []
    for endpoint in declared_operations:
        var_body.append(
            '%sSecurity = map[string]SecurityScheme{' % endpoint.interface_name)

        init_body: List[str] = []
        for scheme, options in endpoint.security.schemes.items():
            init_body.append('"%s": []AuthorizationOption{' % scheme)
            init_body.extend(indent([
                'AuthorizationOption{RequiredScopes: []string{%s}},' % ', '.join(
                    '"{}"'.format(scope) for scope in option.required_scopes)
                for option in options], 1))
            init_body.append('},')
        var_body.extend(indent(init_body, 1))

        var_body.append('}')
    lines.extend(indent(var_body, 1))

    lines.append(')')

    # Declare request & response types for all endpoints
    for endpoint in declared_operations:
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
        body.extend(
            comment(['The result of attempting to authorize this request']))
        body.append('Auth AuthorizationResult')
        lines.extend(indent(body, 1))

        lines.append('}')

        # Declare response type for query
        lines.append('type {} struct {{'.format(endpoint.response_type_name))

        for response in endpoint.responses:
            if response.description:
                lines.extend(
                    indent(comment(response.description.split('\n')), 1))
            body_type = response.json_body_type if response.json_body_type else 'EmptyResponseBody'
            lines.extend(indent(
                ['{} *{}'.format(response.response_set_field, body_type)], 1))
            lines.append('')
        lines.pop()

        lines.append('}')

    lines.append('')
    lines.append('type Implementation interface {')

    body: List[str] = []
    for endpoint in declared_operations:
        comments: List[str] = []
        if endpoint.summary and endpoint.summary != endpoint.description:
            comments.extend(endpoint.summary.split('\n'))
        if endpoint.summary and endpoint.description and endpoint.summary != endpoint.description:
            comments.append('---')
        if endpoint.description:
            comments.extend(endpoint.description.split('\n'))
        body.extend(comment(comments))

        body.append('{}(req *{}) {}'.format(endpoint.interface_name,
                                            endpoint.request_type_name,
                                            endpoint.response_type_name))
        body.append('')
    body.pop()
    lines.extend(indent(body, 1))

    lines.append('}')
    return lines


def routes(declared_operations: List[operations.Operation], path_prefix: Optional[str] = None) -> List[str]:
    """Generate handler Go code for each operation, routed appropriately.

    :param declared_operations: Operations to be included in the handlers and router
    :param path_prefix: Relative path that should be prefixed to every path as declared in the API (e.g., '/scd')
    :return: Lines of Go code defining the handler functions and router creation function
    """
    if path_prefix is None:
        path_prefix = ''
    lines: List[str] = []

    # Define a top-level routed HTTP handler function for each operation
    for operation in declared_operations:
        lines.append(
            'func (s *APIRouter) {}(exp *regexp.Regexp, w http.ResponseWriter, r *http.Request) {{'.format(
                operation.interface_name))

        body: List[str] = []

        # Create object to hold the processed input to the operation
        body.append('var req {}'.format(operation.request_type_name))
        body.append('')

        # Attempt to authorize access to the operation
        body.extend(comment(['Authorize request']))
        body.append(
            'req.Auth = s.Authorizer.Authorize(w, r, &{}Security)'.format(
                operation.interface_name))
        body.append('')

        # Parse any path parameters
        if operation.path_parameters:
            body.extend(comment(['Parse path parameters']))
            body.append('pathMatch := exp.FindStringSubmatch(r.URL.Path)')
            for i, p in enumerate(operation.path_parameters):
                if p.go_type == 'string':
                    body.append(
                        'req.{} = pathMatch[{}]'.format(p.go_field_name, i + 1))
                else:
                    body.append(
                        'req.{} = {}(pathMatch[{}])'.format(p.go_field_name,
                                                            p.go_type, i + 1))
            body.append('')

        # Attempt to parse the request body JSON, if defined
        if operation.json_request_body_type:
            body.extend(comment(['Parse request body']))
            body.append(
                'req.Body = new({})'.format(operation.json_request_body_type))
            body.append('defer r.Body.Close()')
            body.append(
                'req.BodyParseError = json.NewDecoder(r.Body).Decode(req.Body)')
            body.append('')

        # Actually invoke the API Implementation with the processed request to obtain the response
        body.extend(comment(['Call implementation']))
        body.append('response := s.Implementation.{}(&req)'.format(
            operation.interface_name))
        body.append('')

        # Write the first populated response discovered and finish the handler
        body.extend(comment(['Write response to client']))
        responses = [r for r in operation.responses]
        for response in responses:
            body.append(
                'if response.{} != nil {{'.format(response.response_set_field))
            body.extend(indent(['writeJson(w, {}, response.{})'.format(
                response.code, response.response_set_field)], 1))
            body.extend(indent(['return'], 1))
            body.append('}')
        body.append(
            'writeJson(w, 500, InternalServerErrorBody{"Handler implementation did not set a response"})')

        lines.extend(indent(body, 1))

        lines.append('}')
        lines.append('')
    lines.pop()

    lines.append('')

    # Generate a function to create an APIRouter HTTP handler that routes to the appropriate methods in the provided Implementation
    lines.append('func MakeAPIRouter(impl Implementation, auth Authorizer) APIRouter {')

    body: List[str] = []
    body.append(
        'router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*Route, %s)}' % len(
            declared_operations))
    body.append('')
    first_assignment = True
    for i, operation in enumerate(declared_operations):
        path_regex = path_prefix + re.sub(r'{([^}]*)}', r'(?P<\1>[^/]*)',
                                          operation.path)
        body.append('pattern {}= regexp.MustCompile("^{}$")'.format(
            ':' if first_assignment else '', path_regex))
        body.append(
            'router.Routes[%d] = &Route{Pattern: pattern, Handler: router.%s}' % (
            i, operation.interface_name))
        body.append('')
        first_assignment = False
    body.append('return router')
    lines.extend(indent(body, 1))

    lines.append('}')

    return lines


def example(declared_operations: List[operations.Operation]) -> List[str]:
    """Generate Go code for a dummy API Implementation and a main routine to run it.

    Note that this routine produces a starting point/example for implementation,
    but it would be unusual to run it again when updating the API interface.  In
    that case, the data types and interface definitions would be overwritten by
    the new content auto-generated from the updated API interface, but the
    functions initially generated by this routine would generally be manually
    updated rather than being re-auto-generated.

    :param declared_operations: Operations to be included in the Implementation
    :return: Lines of Go code defining a concrete instance of the Implementation interface along with a main function to use it to handle HTTP requests
    """
    lines: List[str] = []

    # Define a concrete instance of the Implementation interface
    lines.append('type ExampleImplementation struct {}')
    lines.append('')
    for endpoint in declared_operations:
        lines.append('func (*ExampleImplementation) {}(req *{}) {} {{'.format(
            endpoint.interface_name, endpoint.request_type_name,
            endpoint.response_type_name))

        body: List[str] = []
        body.append('response := %s{}' % (endpoint.response_type_name))
        # body.append('response.Response500 = &InternalServerErrorBody{ErrorMessage: "Not yet implemented"}')
        body.append('response.%s = &%s{}' % (
            endpoint.responses[0].response_set_field,
            endpoint.responses[0].json_body_type))
        body.append('return response')
        lines.extend(indent(body, 1))

        lines.append('}')
        lines.append('')

    # Define a main function that uses an ExampleImplementation instance to handle HTTP requests
    lines.append('func main() {')

    body: List[str] = []
    body.append(
        'router1 := MakeAPIRouter(&ExampleImplementation{}, &PermissiveAuthorizer{})')
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
