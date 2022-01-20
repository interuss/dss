import dataclasses
from typing import Dict, List, Set, Tuple

import formatting


@dataclasses.dataclass
class ObjectField:
    """A data field within an Object data type"""

    api_name: str
    """Name of the field in the API (generally snake cased)"""

    go_type: str
    """The name of the Go data type which represents this field's value"""

    description: str
    """Documentation of this field"""

    required: bool
    """True if an instance of the parent object must specify a value for this field"""

    @property
    def go_name(self) -> str:
        """Name of the field in the Go representation of the parent data type"""
        return formatting.snake_case_to_pascal_case(self.api_name)


@dataclasses.dataclass
class DataType:
    """A specific data type defined in the API"""

    name: str
    """Name of this data type, as defined in the API"""

    go_type: str = ''
    """Name of the Go data type ('struct' for Object data types)"""

    description: str = ''
    """Documentation of this data type"""

    fields: List[ObjectField] = dataclasses.field(default_factory=list)
    """If this is an Object data type, a list of fields contained in that Object"""

    enum_values: List[str] = dataclasses.field(default_factory=list)
    """If this is a enum data type, a list of values it may take on"""

    def is_primitive(self) -> bool:
        """True iff go_type describes a built-in Go type"""
        return self.go_type in go_primitives.values() or self.go_type in go_numbers.values()


go_primitives: Dict[str, str] = {
    'string': 'string',
    'boolean': 'bool',
}
"""Maps OpenAPI `type` to Go primitive type"""

go_numbers: Dict[str, str] = {
    'float': 'float32',
    'double': 'float64',
    'int32': 'int32',
    'int64': 'int64',
    'number': 'float64',
    'integer': 'float32',
}
"""Maps OpenAPI `format` (defaulting to `type` if `format` is missing) to Go primitive type"""


def get_data_type_name(component_name: str, data_type_name: str) -> str:
    """Get the plain data type name from a $ref URI.

    :param component_name: $ref URI to the data type of interest
    :param data_type_name: context in which the data type is being retrieved (used for error message only)
    :return: Plain data type name in the relative $ref URI
    """
    if component_name == '':
        return ''
    elif component_name.startswith('#/components/schemas/'):
        return component_name[len('#/components/schemas/'):]
    else:
        raise NotImplementedError('$ref expected to start with `#/components/schemas/`, but found `{}` instead for {}'.format(component_name, data_type_name))


def _parse_referenced_type_name(schema: Dict, data_type_name: str) -> str:
    options = schema['anyOf'] if 'anyOf' in schema else schema['allOf']
    if len(options) != 1:
        raise NotImplementedError('Only one $ref is supported for anyOf and allOf; found {} elements instead'.format(len(options)))
    option = options[0]
    if not isinstance(option, dict):
        raise ValueError('Expected dict entries in anyOf/allOf block; found {} instead'.format(option))
    if len(option) != 1 or '$ref' not in option:
        raise NotImplementedError('The only element in anyOf/allOf must be a $ref dictionary; found {} instead'.format(option))
    return get_data_type_name(option['$ref'], data_type_name)


def make_object_field(object_name: str, field_name: str, schema: Dict, required: Set[str]) -> Tuple[ObjectField, List[DataType]]:
    """Parse a single field in a data type or endpoint parameter schema.

    :param object_name: Name of the object containing this field, for error messages and inline type names
    :param field_name: Name of the object field being parsed
    :param schema: Definition of the object field being parsed
    :param required: The set of required fields for the parent object
    :return: Tuple of
      * The object field defined by the provided schema
      * Any additional data types incidentally defined in the provided schema
    """
    is_required = field_name in required
    if '$ref' in schema:
        return ObjectField(
            api_name=field_name,
            go_type=get_data_type_name(schema['$ref'], object_name),
            description=schema.get('description', ''),
            required=is_required), []
    elif 'anyOf' in schema or 'allOf' in schema:
        return ObjectField(
            api_name=field_name,
            go_type=_parse_referenced_type_name(schema, object_name + '.' + field_name),
            description=schema.get('description', ''),
            required=is_required), []
    else:
        type_name = object_name + formatting.snake_case_to_pascal_case(field_name)
        data_type, additional_types = make_data_types(type_name, schema)
        if data_type.go_type in go_primitives.values():
            # No additional type declaration needed
            if additional_types:
                raise RuntimeError('{} field type `{}` was parsed as primitive {} but also generated {} additional types'.format(object_name, field_name, data_type.go_type, len(additional_types)))
            field_data_type = data_type.go_type
        else:
            if data_type.go_type.startswith('[]'):
                field_data_type = data_type.go_type
            else:
                additional_types.append(data_type)
                field_data_type = data_type.name
        return ObjectField(
            api_name=field_name,
            go_type=field_data_type,
            description=data_type.description,
            required=is_required), additional_types


def _make_object_fields(object_name: str, properties: Dict, required: Set[str]) -> Tuple[List[ObjectField], List[DataType]]:
    fields: List[ObjectField] = []
    additional_types: List[DataType] = []
    for field_name, schema in properties.items():
        field, further_types = make_object_field(object_name, field_name, schema, required)
        additional_types.extend(further_types)
        fields.append(field)
    return fields, additional_types


def make_data_types(name: str, schema: Dict) -> Tuple[DataType, List[DataType]]:
    """Parse all data types necessary to express the provided data type schema.

    In addition to the primary data type described by `name`, this routine also
    generates additional data types defined inline in the provided schema.

    :param name: Name of the primary data type being parsed
    :param schema: Definition of the data type being parsed
    :return: Tuple of
      * The primary data defined by the provided schema
      * Any additional data types incidentally defined in the provided schema
    """
    data_type = DataType(name=name)
    additional_types = []

    if 'description' in schema:
        data_type.description = schema['description']

    if 'type' in schema:
        if schema['type'] in go_primitives:
            data_type.go_type = go_primitives[schema['type']]
        elif schema['type'] in {'number', 'integer'}:
            data_type.go_type = go_numbers.get(schema.get('format', schema['type']), '')
            if not data_type.go_type:
                raise ValueError('Unrecognized numeric format `{}` for {}'.format(schema.get('format', '<missing>'), name))
        elif schema['type'] == 'array':
            if 'items' in schema:
                items = schema['items']
                if '$ref' in items:
                    data_type.go_type = '[]{}'.format(get_data_type_name(items['$ref'], name))
                else:
                    item_type, further_types = make_data_types(name + 'Item', items)
                    additional_types.extend(further_types)
                    if item_type.description != '' or not item_type.is_primitive():
                        additional_types.append(item_type)
                        data_type.go_type = '[]{}'.format(item_type.name)
                    else:
                        data_type.go_type = '[]{}'.format(item_type.go_type)
            else:
                raise ValueError('Missing `items` declaration for {} array type'.format(name))
        elif schema['type'] == 'object':
            data_type.go_type = 'struct'
            data_type.fields, further_types = _make_object_fields(
                name,
                schema.get('properties', {}),
                set(schema.get('required', [])))
            additional_types.extend(further_types)
        else:
            raise ValueError('Unrecognized type `{}` in {} type'.format(schema['type'], name))
    elif 'anyOf' in schema or 'allOf' in schema:
        data_type.go_type = _parse_referenced_type_name(schema, name)

    if 'enum' in schema:
        data_type.enum_values = schema['enum']

    return data_type, additional_types
