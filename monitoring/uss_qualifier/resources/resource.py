from abc import ABC, abstractmethod
import inspect
from typing import Dict, Generic, TypeVar

from implicitdict import ImplicitDict

from monitoring import uss_qualifier as uss_qualifier_module
from monitoring.monitorlib import inspection
from monitoring.uss_qualifier import resources as resources_module
from monitoring.uss_qualifier.resources.definitions import (
    ResourceDeclaration,
    ResourceID,
)

SpecificationType = TypeVar("SpecificationType", bound=ImplicitDict)


class Resource(ABC, Generic[SpecificationType]):
    @abstractmethod
    def __init__(self, specification: SpecificationType, **dependencies):
        """Create an instance of the resource.

        Concrete subclasses of Resource must implement their constructor according to this specification.

        :param specification: A serializable (subclass of implicitdict.ImplicitDict) specification for how to create the resource.
        :param dependencies: If this resource depends on any other resources, each of the other dependencies should be declared as an additional typed parameter to the constructor.  Each parameter type should be a class that is a subclass of Resource.
        """
        raise NotImplementedError(
            "A concrete resource type must implement __init__ method"
        )

    def is_type(self, resource_type: str) -> bool:
        specified_type = inspection.get_module_object_by_name(
            uss_qualifier_module, resource_type
        )
        return self.__class__ == specified_type


ResourceType = TypeVar("ResourceType", bound=Resource)


def create_resources(
    resource_declarations: Dict[ResourceID, ResourceDeclaration]
) -> Dict[ResourceID, ResourceType]:
    resource_pool: Dict[ResourceID, ResourceType] = {}

    resources_created = 1
    unmet_dependencies_by_resource = {}
    while resources_created > 0:
        resources_created = 0
        for name, declaration in resource_declarations.items():
            if name in resource_pool:
                continue
            unmet_dependencies = [
                d for d in declaration.dependencies.values() if d not in resource_pool
            ]
            if unmet_dependencies:
                unmet_dependencies_by_resource[name] = unmet_dependencies
            else:
                resource_pool[name] = _make_resource(declaration, resource_pool)
                resources_created += 1

    if len(resource_pool) != len(resource_declarations):
        uncreated_resources = [
            (r + " ({} missing)".format(", ".join(unmet_dependencies_by_resource[r])))
            for r in resource_declarations
            if r not in resource_pool
        ]
        raise ValueError(
            "Could not create resources: {} (do you have circular dependencies?)".format(
                ", ".join(uncreated_resources)
            )
        )

    return resource_pool


def _make_resource(
    declaration: ResourceDeclaration, resource_pool: Dict[ResourceID, Resource]
) -> Resource:
    inspection.import_submodules(resources_module)
    resource_type = inspection.get_module_object_by_name(
        uss_qualifier_module, declaration.resource_type
    )
    if not issubclass(resource_type, Resource):
        raise NotImplementedError(
            "Resource type {} is not a subclass of the Resource base class".format(
                resource_type.__name__
            )
        )

    constructor_signature = inspect.signature(resource_type.__init__)
    specification_type = None
    constructor_args = {}
    for arg_name, arg in constructor_signature.parameters.items():
        if arg_name == "self":
            continue
        if arg_name == "specification":
            specification_type = arg.annotation
            continue
        if arg_name not in declaration.dependencies:
            raise ValueError(
                'Resource declaration for {} is missing a source for dependency "{}"'.format(
                    declaration.resource_type, arg
                )
            )
        if declaration.dependencies[arg_name] not in resource_pool:
            raise ValueError(
                'Resource "{}" was not found in the resource pool when trying to create {} resource'.format(
                    declaration.dependencies[arg_name], declaration.resource_type
                )
            )
        constructor_args[arg_name] = resource_pool[declaration.dependencies[arg_name]]
    if specification_type is not None:
        constructor_args["specification"] = ImplicitDict.parse(
            declaration.specification, specification_type
        )

    return resource_type(**constructor_args)
