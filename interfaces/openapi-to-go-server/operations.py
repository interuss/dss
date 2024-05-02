import dataclasses
from typing import Dict, List, Tuple

import data_types
import formatting


@dataclasses.dataclass
class StringParameter:
    """Single parameter to an operation found in the path used to invoke the operation"""

    name: str
    """Name of this parameter"""

    description: str
    """Documentation for this parameter"""

    go_type: str
    """The Go data type that holds the value of this parameter"""

    @property
    def go_field_name(self) -> str:
        """Go-style field name for this parameter in the Operation's `request_type_name`"""
        return formatting.snake_case_to_pascal_case(self.name)


@dataclasses.dataclass
class Scope:
    """Representation of a scope used to authorize a particular Operation"""

    name: str
    """Name of the scope"""

    @property
    def go_constant_name(self) -> str:
        """Go-style field name for this parameter in the Operation's `request_type_name`"""
        return "{}Scope".format(formatting.string_to_pascal_case(self.name))

    def __hash__(self):
        return self.name.__hash__()


@dataclasses.dataclass
class AuthorizationOption:
    """One acceptable option for authorization to invoke a particular Operation"""

    option: Dict[str, List[Scope]]
    """Mapping between authorization scheme name and a list of scope combination options that may be used to access the Operation under that authorization scheme"""


@dataclasses.dataclass
class Security:
    """Set of all defined authorization options for a particular Operation"""

    options: List[AuthorizationOption]
    """List of authorization options that may be used to access the Operation"""


@dataclasses.dataclass
class Response:
    """One possible response from an Operation, as defined in the API"""

    code: int
    """HTTP response code for this response"""

    description: str
    """Documentation regarding when this response is returned"""

    json_body_type: str
    """Response body type, if an application/json response body is defined for this response (blank otherwise)"""

    @property
    def response_set_field(self) -> str:
        """Name of the field in the Operation's `response_type_name` where this response can be set"""
        return 'Response{}'.format(self.code)


@dataclasses.dataclass
class Operation:
    """An operation uniquely identified with a path and HTTP verb"""

    path: str
    """Path of this operation relative to the base URL"""

    summary: str
    """Summary documentation for this operation or path"""

    description: str
    """Description documentation for this operation or path"""

    operation_id: str
    """Explicitly defined operation ID for this operation"""

    tags: List[str]
    """Set of tags which apply to this operation"""

    security: Security
    """Set of security schemes that may be used to access this operation"""

    verb: str
    """HTTP verb for this operation"""

    path_parameters: List[StringParameter]
    """Parameters found in the path when invoking this operation"""

    query_parameters: List[StringParameter]
    """Parameters found in the query when invoking this operation"""

    json_request_body_type: str
    """Request body type, if an application/json request body is defined for this operation (blank otherwise)"""

    responses: List[Response]
    """All defined responses that may be returned from this operation"""

    @property
    def interface_name(self) -> str:
        """Go-style name of this operation, as would appear in an interface"""
        if self.operation_id:
            return formatting.capitalize_first_letter(self.operation_id)
        else:
            return formatting.capitalize_first_letter(
                self.verb.lower()) + formatting.snake_case_to_pascal_case(
                self.path.replace('{', '').replace('}', '').replace('/', '_'))

    @property
    def response_type_name(self) -> str:
        """Name of the Go type that contains all of the defined responses"""
        return self.interface_name + 'ResponseSet'

    @property
    def request_type_name(self) -> str:
        """Name of the Go type that contains the non-body request parameters"""
        return self.interface_name + 'Request'

    @property
    def verb_const_name(self) -> str:
        """Name of the Go const that contains the HTTP verb"""
        return 'http.Method' + formatting.capitalize_first_letter(self.verb.lower())

def _parse_parameters(schema: Dict) -> Tuple[List[StringParameter], List[StringParameter], List[data_types.DataType]]:
    """Parse operation parameters from an OpenAPI schema for path or verb

    :param schema: Schema for an OpenAPI path or an OpenAPI verb
    :return: Parameters discovered including:
      * Path parameters
      * Query parameters
    """
    path_parameters: List[StringParameter] = []
    query_parameters: List[StringParameter] = []
    additional_types: List[data_types.DataType] = []
    for parameter in schema.get('parameters', []):
        parameter_name = parameter['name']
        parameter_description = parameter.get('description', '')
        parameter_in = parameter['in']
        if 'schema' in parameter:
            parameter_field, further_types = data_types.make_object_field(
                '', parameter_name, parameter['schema'], set())
            if further_types:
                additional_types.extend(further_types)
            parameter_type = parameter_field.go_type
        else:
            parameter_type = 'string'
        if parameter_in == 'path':
            path_parameters.append(
                StringParameter(name=parameter_name,
                                description=parameter_description,
                                go_type=parameter_type))
        elif parameter_in == 'query':
            query_parameters.append(
                StringParameter(name=parameter_name,
                                description=parameter_description,
                                go_type=parameter_type))
        else:
            raise NotImplementedError(
                'Parameter in "{}" (`{}`) not yet implemented'.format(
                    parameter_in,
                    parameter_name))

    return path_parameters, query_parameters, additional_types


def make_operations(path: str, schema: Dict) -> Tuple[List[Operation], List[data_types.DataType]]:
    """Parse all operations defined within the specified path definition.

    :param path: Relative path of operations described in `schema`
    :param schema: Definition of operations accessible at `path`
    :return:
        * Operations defined by `schema`
        * Additional data types implicitly defined within operation definitions
    """
    declared_operations: List[Operation] = []

    summary = schema.get('summary', '')
    description = schema.get('description', '')

    # Parse common parameters for all operations in schema
    common_path_parameters, common_query_parameters, additional_data_types = _parse_parameters(schema)

    # Parse each operation defined in schema
    for verb in ('get', 'put', 'post', 'delete'):
        if verb not in schema:
            continue
        action = schema[verb]
        verb_summary = action.get('summary', summary)
        verb_description = action.get('description', description)
        operation_id = action.get('operationId', '')
        tags = action.get('tags', [])
        component_name = action.get('requestBody', {}).get('content', {}).get('application/json', {}).get('schema', {}).get('$ref', '')
        request_body_type = data_types.get_data_type_name(component_name, 'requestBody')

        path_parameters, query_parameters, further_data_types = _parse_parameters(action)
        path_parameters += common_path_parameters
        query_parameters += common_query_parameters
        additional_data_types.extend(further_data_types)

        security = Security(options=[])
        for security_option in action.get('security', []):
            auth_option = AuthorizationOption(option={})
            for scheme, scopes in security_option.items():
                auth_option.option[scheme] = [Scope(s) for s in scopes]
            security.options.append(auth_option)

        responses: List[Response] = []
        for code, response in action.get('responses', {}).items():
            component_name = response.get('content', {}).get('application/json', {}).get('schema', {}).get('$ref', '')
            json_body_type = data_types.get_data_type_name(component_name, 'response body')
            responses.append(Response(
                code=int(code),
                description=response.get('description', ''),
                json_body_type=json_body_type))

        declared_operations.append(Operation(
            path=path,
            summary=verb_summary,
            description=verb_description,
            operation_id=operation_id,
            tags=tags,
            security=security,
            verb=verb,
            path_parameters=path_parameters,
            query_parameters=query_parameters,
            json_request_body_type=request_body_type,
            responses=responses
        ))

    return declared_operations, additional_data_types
