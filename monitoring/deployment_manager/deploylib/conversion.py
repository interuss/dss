import re
from typing import Any

import kubernetes


def to_openapi_object(value: Any, type_name: str) -> Any:
    primitives = {'int', 'str'}

    # Do not convert primitives
    if type_name in primitives:
        return value

    # Convert dicts
    dict_regex = r'dict\(([^,]+), ([^,]+)\)'
    m = re.fullmatch(dict_regex, type_name)
    if m:
        return {to_openapi_object(k, m.group(1)): to_openapi_object(v, m.group(2)) for k, v in value.items()}

    # Convert lists
    list_regex = r'list\[([^]]+)\]'
    m = re.fullmatch(list_regex, type_name)
    if m:
        return [to_openapi_object(v, m.group(1)) for v in value]

    # Convert OpenAPI objects
    api_type = getattr(kubernetes.client, type_name)
    kwargs = {}
    api_arg_names = api_type.attribute_map
    for arg_name, arg_type_name in api_type.openapi_types.items():
        api_arg_name = api_arg_names[arg_name]
        if api_arg_name in value:
            kwargs[arg_name] = to_openapi_object(value[api_arg_name], arg_type_name)
    return api_type(**kwargs)
