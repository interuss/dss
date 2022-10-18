from typing import Dict, TypeVar

from implicitdict import ImplicitDict


ResourceID = str
"""This plain string represents the ID/name of a resource"""


ResourceTypeName = str
"""This plain string represents a type of resource, expressed as a Python class name qualified relative to this `resources` module"""


SpecificationType = TypeVar("SpecificationType", bound=ImplicitDict)


class ResourceDeclaration(ImplicitDict):
    resource_type: ResourceTypeName
    """Type of resource, expressed as a Python class name qualified relative to this `resources` module"""

    dependencies: Dict[ResourceID, ResourceID] = {}
    """Mapping of dependency parameter (additional argument to concrete resource constructor) to `name` of resource to use"""

    specification: dict = {}
    """Specification of resource; format is the SpecificationType that corresponds to the `resource_type`"""


class ResourceCollection(ImplicitDict):
    resource_declarations: Dict[ResourceID, ResourceDeclaration]
    """Mapping of globally (within resource collection) unique name identifying a resource to the declaration of that resource"""
