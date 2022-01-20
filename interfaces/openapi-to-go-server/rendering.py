import re
from typing import Dict, List

import apis
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


def imports(import_list: List[str]) -> str:
    return '\n'.join(indent(['"{}"'.format(i) for i in import_list], 2))


def template_content(template_name: str, template_vars: Dict[str, str]) -> str:
    """Fill in a template with provided values and return the entire content.

    :param template_name: Name of template file in `templates` folder (e.g., 'common' reads from `templates/common.go.template`)
    :param template_vars: Mapping of key (sentinel in template) to value (what to replace the sentinel with)
    :return: Template content with filled values
    """
    with open('templates/{}.go.template'.format(template_name), 'r') as f:
        content = f.read()
    for k, v in template_vars.items():
        content = content.replace(k, v)
    return content


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


def implementation_interface(api: apis.API, api_package: str, ensure_500: bool) -> List[str]:
    """Generate Go code defining the interface an API implementation must implement.

    :param api: API to be rendered into an interface
    :param api_package: Name of root/common API package
    :param ensure_500: If True, add a 500 response to all operations that don't already define a 500 response
    :return: Lines of Go code defining the interface
    """
    lines: List[str] = []

    # Provide security constants
    lines.append('var (')

    var_body: List[str] = []
    for operation in api.operations:
        var_body.append(
            '%sSecurity = map[string]%s.SecurityScheme{' % (operation.interface_name, api_package))

        init_body: List[str] = []
        for scheme, options in operation.security.schemes.items():
            init_body.append('"%s": []%s.AuthorizationOption{' % (scheme, api_package))
            init_body.extend(indent([
                '%s.AuthorizationOption{RequiredScopes: []string{%s}},' % (
                    api_package,
                    ', '.join('"{}"'.format(scope)
                              for scope in option.required_scopes))
                for option in options], 1))
            init_body.append('},')
        var_body.extend(indent(init_body, 1))

        var_body.append('}')
    lines.extend(indent(var_body, 1))

    lines.append(')')

    # Declare request & response types for all operations
    for operation in api.operations:
        lines.append('')

        # Declare request type for operation
        lines.append('type {} struct {{'.format(operation.request_type_name))

        body: List[str] = []
        for p in operation.path_parameters + operation.query_parameters:
            if p.description:
                body.extend(comment(p.description.split('\n')))
            body.append('{} {}'.format(p.go_field_name, p.go_type))
            body.append('')
        if operation.json_request_body_type:
            body.extend(comment(['The data contained in the body of this request, if it parsed correctly']))
            body.append('Body *{}'.format(operation.json_request_body_type))
            body.append('')
            body.extend(comment(['The error encountered when attempting to parse the body of this request']))
            body.append('BodyParseError error')
            body.append('')
        body.extend(
            comment(['The result of attempting to authorize this request']))
        body.append('Auth {}.AuthorizationResult'.format(api_package))
        lines.extend(indent(body, 1))

        lines.append('}')

        # Declare response type for operation
        lines.append('type {} struct {{'.format(operation.response_type_name))

        body: List[str] = []
        responses = [r for r in operation.responses]
        if ensure_500 and 500 not in {r.code for r in responses}:
            responses.append(operations.Response(code=500, description='Auto-generated internal server error response', json_body_type='{}.InternalServerErrorBody'.format(api_package)))
        for response in responses:
            if response.description:
                body.extend(comment(response.description.split('\n')))
            body_type = response.json_body_type if response.json_body_type else 'EmptyResponseBody'
            body.extend(['{} *{}'.format(response.response_set_field, body_type)])
            body.append('')
        body.pop()
        lines.extend(indent(body, 1))

        lines.append('}')

    lines.append('')
    lines.append('type Implementation interface {')

    body: List[str] = []
    for operation in api.operations:
        comments: List[str] = []
        if operation.summary and operation.summary != operation.description:
            comments.extend(operation.summary.split('\n'))
        if operation.summary and operation.description and operation.summary != operation.description:
            comments.append('---')
        if operation.description:
            comments.extend(operation.description.split('\n'))
        body.extend(comment(comments))

        body.append('{}(req *{}) {}'.format(operation.interface_name,
                                            operation.request_type_name,
                                            operation.response_type_name))
        body.append('')
    body.pop()
    lines.extend(indent(body, 1))

    lines.append('}')
    return lines


def routes(api: apis.API, api_package: str, ensure_500: bool) -> List[str]:
    """Generate handler Go code for each operation, routed appropriately.

    :param api: API to have its operation routes rendered
    :param api_package: Name of root/common API package
    :param ensure_500: If True, add a 500 response to all operations that don't already define a 500 response
    :return: Lines of Go code defining the handler functions and router creation function
    """
    lines: List[str] = []

    # Define a top-level routed HTTP handler function for each operation
    for operation in api.operations:
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

        # Parse any query parameters
        # TODO

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
        if ensure_500 and 500 not in {r.code for r in responses}:
            responses.append(operations.Response(code=500, description='', json_body_type='{}.InternalServerErrorBody'.format(api_package)))
        for response in responses:
            body.append(
                'if response.{} != nil {{'.format(response.response_set_field))
            body.extend(indent(['{}.WriteJson(w, {}, response.{})'.format(
                api_package, response.code, response.response_set_field)], 1))
            body.extend(indent(['return'], 1))
            body.append('}')
        body.append('%s.WriteJson(w, 500, %s.InternalServerErrorBody{"Handler implementation did not set a response"})' % (api_package, api_package))

        lines.extend(indent(body, 1))

        lines.append('}')
        lines.append('')
    lines.pop()

    return lines


def routing(api: apis.API, api_package: str) -> List[str]:
    """Generate Go code to create an APIRouter for the provided Implementation.

    :param api: API to have its operation routes rendered
    :param api_package: Name of root/common API package
    :return: Lines of Go code defining the contents of a function to create an APIRouter that routes to the appropriate methods in the provided Implementation
    """
    lines: List[str] = []
    lines.append(
        'router := APIRouter{Implementation: impl, Authorizer: auth, Routes: make([]*%s.Route, %d)}' % (api_package, len(api.operations)))
    lines.append('')
    first_assignment = True
    for i, operation in enumerate(api.operations):
        path_regex = '/' + api.package + re.sub(r'{([^}]*)}', r'(?P<\1>[^/]*)',
                                          operation.path)
        lines.append('pattern {}= regexp.MustCompile("^{}$")'.format(
            ':' if first_assignment else '', path_regex))
        lines.append(
            'router.Routes[%d] = &%s.Route{Pattern: pattern, Handler: router.%s}' % (
            i, api_package, operation.interface_name))
        lines.append('')
        first_assignment = False
    lines.append('return router')
    return lines


def example_implementation(api: apis.API, implementation_name: str) -> List[str]:
    """Generate Go code for a dummy API Implementation and a main routine to run it.

    Note that this routine produces a starting point/example for implementation,
    but it would be unusual to run it again when updating the API interface.  In
    that case, the data types and interface definitions would be overwritten by
    the new content auto-generated from the updated API interface, but the
    functions initially generated by this routine would generally be manually
    updated rather than being re-auto-generated.

    :param api: API to have its operation routes rendered
    :return: Lines of Go code defining a concrete instance of the Implementation interface
    """
    lines: List[str] = []

    # Define a concrete instance of the Implementation interface
    lines.append('type %s struct {}' % implementation_name)
    lines.append('')
    for operation in api.operations:
        lines.append('func (*{}) {}(req *{}) {} {{'.format(
            implementation_name, operation.interface_name,
            api.package + '.' + operation.request_type_name,
            api.package + '.' + operation.response_type_name))

        body: List[str] = []
        body.append('response := %s{}' % (api.package + '.' + operation.response_type_name))
        # body.append('response.Response500 = &InternalServerErrorBody{ErrorMessage: "Not yet implemented"}')
        body.append('response.%s = &%s{}' % (
            operation.responses[0].response_set_field,
            api.package + '.' + operation.responses[0].json_body_type))
        body.append('return response')
        lines.extend(indent(body, 1))

        lines.append('}')
        lines.append('')

    return lines


def example_router_defs(implementations: Dict[str, str], api_package: str) -> List[str]:
    """Generate Go code for concrete example router & multi-router definitions.

    :param implementations: Relationship between API name and the name of the Go struct implementing the Implementation of that API
    :param api_package: Name of root/common API package
    :return: Lines of Go code for router definitions in a main function to use to handle HTTP requests
    """
    lines: List[str] = []

    for api_name, implementation in implementations.items():
        lines.append('%sRouter := %s.MakeAPIRouter(&%s{}, &authorizer)' % (api_name, api_name, implementation))
    router_list = ', '.join('&{}Router'.format(api_name) for api_name, _ in implementations.items())
    lines.append('multiRouter := %s.MultiRouter{Routers: []%s.APIRouter{%s}}' % (api_package, api_package, router_list))

    return lines
