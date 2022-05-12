from typing import Any, Callable, Dict, List, Optional, Type

from kubernetes.client import V1ObjectMeta, V1Deployment, V1Ingress, V1Namespace, V1Service, V1Secret


_special_comparisons: Dict[Type, Callable[[Any, Any], bool]] = {}


def specs_are_the_same(obj1: Any, obj2: Any, field_paths: Optional[List[str]]=None) -> bool:
    """Determine if the specifications for two Kubernetes objects are equivalent

    Only the relevant parts of the two objects will be compared to see if they
    are describing the same specification.  The relevant descendant fields are
    described by field_paths (if specified).  If field_paths is not specified or
    empty, then first the _special_comparisons registry will be checked to see
    if there is a custom comparison function for obj1's class type.  If not,
    field_paths will be set to the list of keys in obj1.attribute_map if it
    exists.  Finally, obj1 and obj2 will be checked for simple equality if none
    of the above is applicable.

    :param obj1: Kubernetes object or field value
    :param obj2: Kubernetes object or field value
    :param field_paths: Specific descendant paths to check, if specified
    :return: Whether the specs described by the two objects are the same
    """
    if not field_paths:
        special_are_the_same = _special_comparisons.get(obj1.__class__, None)
        if special_are_the_same is not None:
            return special_are_the_same(obj1, obj2)
        elif hasattr(obj1, 'attribute_map'):
            return specs_are_the_same(obj1, obj2, list(getattr(obj1, 'attribute_map').keys))
        elif isinstance(obj1, dict) and isinstance(obj2, dict):
            return all(obj2[k] == v for k, v in obj1.items()) and all(obj1[k] == v for k, v in obj2.items())
        elif isinstance(obj1, list) and isinstance(obj2, list):
            return all(v1 == v2 for v1, v2 in zip(obj1, obj2))
        else:
            return obj1 == obj2

    sub_paths: Dict[str, Optional[List[str]]] = {}
    for field_path in field_paths:
        parts = field_path.split('.')
        if len(parts) == 1:
            if parts[0] in sub_paths:
                raise ValueError('Cannot compare {} and its subfield {}'.format(parts[0], field_path))
            sub_paths[parts[0]] = None
        else:
            sub_fields = sub_paths.get(parts[0], [])
            sub_fields.append('.'.join(parts[1:]))
            sub_paths[parts[0]] = sub_fields

    for field_name, paths in sub_paths.items():
        if not hasattr(obj1, field_name) and not hasattr(obj2, field_name):
            continue
        elif hasattr(obj1, field_name) and hasattr(obj2, field_name):
            value1 = getattr(obj1, field_name)
            value2 = getattr(obj2, field_name)
            if not specs_are_the_same(value1, value2, paths):
                return False
        else:
            return False

    return True


def _special_comparison(type: Type):
    def decorator_declare_comparison(compare: Callable[[Any, Any], bool]):
        global _special_comparisons
        _special_comparisons[type] = compare
        return compare
    return decorator_declare_comparison


@_special_comparison(V1Namespace)
def _v1namespace_are_the_same(namespace1: V1Namespace, namespace2: V1Namespace) -> bool:
    return specs_are_the_same(namespace1, namespace2, ['metadata'])


@_special_comparison(V1ObjectMeta)
def _v1objectmeta_are_the_same(meta1: V1ObjectMeta, meta2: V1ObjectMeta) -> bool:
    return specs_are_the_same(meta1, meta2, ['annotations', 'labels', 'name', 'namespace'])


@_special_comparison(V1Deployment)
def _v1deployment_are_the_same(dep1: V1Deployment, dep2: V1Deployment) -> bool:
    return specs_are_the_same(dep1, dep2, ['metadata', 'spec'])


@_special_comparison(V1Service)
def _v1service_are_the_same(svc1: V1Service, svc2: V1Service) -> bool:
    return specs_are_the_same(svc1, svc2, ['metadata', 'spec'])


@_special_comparison(V1Ingress)
def _v1ingress_are_the_same(ingress1: V1Ingress, ingress2: V1Ingress) -> bool:
    return specs_are_the_same(ingress1, ingress2, ['metadata', 'spec'])


@_special_comparison(V1Secret)
def _v1secret_are_the_same(secret1: V1Secret, secret2: V1Secret) -> bool:
    return specs_are_the_same(secret1, secret2, ['metadata', 'data'])
