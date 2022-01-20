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
        self.operations = [op for op in self.operations
                           if not all(tag not in op.tags for tag in tags)]
        # TODO: Remove unnecessary DataTypes


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
