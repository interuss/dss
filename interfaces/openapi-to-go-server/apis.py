import dataclasses
from typing import Dict, List, Set

import data_types
import operations


@dataclasses.dataclass
class API:
    package: str
    data_types: List[data_types.DataType]
    operations: List[operations.Operation]

    def filter_operations(self, tags: Set[str]):
        # Select only the applicable operations
        self.operations = [op for op in self.operations
                           if not all(tag not in op.tags for tag in tags)]

        # Determine the necessary data types
        required_data_types: Set[str] = set()
        for op in self.operations:
            for p in op.path_parameters + op.query_parameters:
                required_data_types.add(p.go_type)
            if op.json_request_body_type:
                required_data_types.add(op.json_request_body_type)
            for response in op.responses:
                if response.json_body_type:
                    required_data_types.add(response.json_body_type)

        data_types_by_name: Dict[str, data_types.DataType] = {dt.name: dt for dt in self.data_types}
        data_types_to_check = [dt for dt in required_data_types]
        while data_types_to_check:
            data_type_name = data_types_to_check.pop()
            base_data_type = data_type_name[2:] if data_type_name.startswith('[]') else data_type_name
            required_data_types.add(base_data_type)
            data_type = data_types_by_name.get(base_data_type, None)
            if not data_type:
                continue
            if data_type.go_type not in required_data_types:
                required_data_types.add(data_type.go_type)
                data_types_to_check.append(data_type.go_type)
            for field in data_type.fields:
                if field.go_type not in required_data_types:
                    required_data_types.add(field.go_type)
                    data_types_to_check.append(field.go_type)

        # Select only the necessary data types
        self.data_types = [dt for dt in self.data_types
                           if dt.name in required_data_types]


def make_api(package: str, spec: Dict) -> API:
    # Parse all defined data types
    if 'components' not in spec:
        raise ValueError('Missing `components` in YAML for {}'.format(package))
    components = spec['components']
    if 'schemas' not in components:
        raise ValueError('Missing `schemas` in `components` for {}'.format(package))
    declared_types = []
    for name, schema in components['schemas'].items():
        data_type, additional_types = data_types.make_data_types(name, schema)
        declared_types.extend(additional_types)
        declared_types.append(data_type)

    # Parse all endpoints
    if 'paths' not in spec:
        raise ValueError('Missing `paths` in YAML for {}'.format(package))
    paths = spec['paths']
    declared_operations = []
    for name, schema in paths.items():
        new_operations = operations.make_operations(name, schema)
        declared_operations.extend(new_operations)

    return API(package=package, data_types=declared_types, operations=declared_operations)
