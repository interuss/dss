import dataclasses
from typing import Dict, List, Set, Tuple


def snake_case_to_pascal_case(s: str) -> str:
    return s.replace('_', ' ').title().replace(' ', '')


@dataclasses.dataclass
class ObjectField:
    api_name: str
    go_type: str
    description: str
    required: bool

    @property
    def go_name(self) -> str:
        return snake_case_to_pascal_case(self.api_name)


@dataclasses.dataclass
class DataType:
    name: str
    go_type: str = ''
    description: str = ''
    fields: List[ObjectField] = dataclasses.field(default_factory=list)
    enum_values: List[str] = dataclasses.field(default_factory=list)

    def is_primitive(self) -> bool:
        return self.go_type in go_primitives.values() or self.go_type in go_numbers.values()


# Maps OpenAPI `type` to Go primitive type
go_primitives: Dict[str, str] = {
    'string': 'string',
    'boolean': 'bool',
}

# Maps OpenAPI `format` (defaulting to `type` if `format` is missing) to Go primitive type
go_numbers: Dict[str, str] = {
    'float': 'float32',
    'double': 'float64',
    'int32': 'int32',
    'int64': 'int64',
    'number': 'float64',
    'integer': 'float32',
}


def get_data_type_name(component_name: str, data_type_name: str) -> str:
    if component_name == '':
        return ''
    elif component_name.startswith('#/components/schemas/'):
        return component_name[len('#/components/schemas/'):]
    else:
        raise NotImplementedError('$ref expected to start with `#/components/schemas/`, but found `{}` instead for {}'.format(component_name, data_type_name))


def parse_referenced_type_name(schema: Dict, data_type_name: str) -> str:
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
    if '$ref' in schema:
        return ObjectField(api_name=field_name, go_type=get_data_type_name(schema['$ref'], object_name), description=schema.get('description', ''), required=field_name in required), []
    elif 'anyOf' in schema or 'allOf' in schema:
        return ObjectField(api_name=field_name, go_type=parse_referenced_type_name(schema, object_name + '.' + field_name), description=schema.get('description', ''), required=field_name in required), []
    else:
        data_type, additional_types = make_data_types(object_name + snake_case_to_pascal_case(field_name), schema)
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
        return ObjectField(api_name=field_name, go_type=field_data_type, description=data_type.description, required=field_name in required), additional_types


def make_object_fields(object_name: str, properties: Dict, required: Set[str]) -> Tuple[List[ObjectField], List[DataType]]:
    fields: List[ObjectField] = []
    additional_types: List[DataType] = []
    for field_name, schema in properties.items():
        field, further_types = make_object_field(object_name, field_name, schema, required)
        additional_types.extend(further_types)
        fields.append(field)
    return fields, additional_types


def make_data_types(name: str, schema: Dict) -> Tuple[DataType, List[DataType]]:
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
            data_type.fields, further_types = make_object_fields(name, schema.get('properties', {}), set(schema.get('required', [])))
            additional_types.extend(further_types)
        else:
            raise ValueError('Unrecognized type `{}` in {} type'.format(schema['type'], name))
    elif 'anyOf' in schema or 'allOf' in schema:
        data_type.go_type = parse_referenced_type_name(schema, name)

    if 'enum' in schema:
        data_type.enum_values = schema['enum']

    return data_type, additional_types
