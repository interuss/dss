from typing import Any, Callable, Optional

from structlog import BoundLogger

from monitoring.deployment_manager.deploylib import comparisons


def get_resource(list_resources: Callable[[], Any], log: BoundLogger, resource_type: str, resource_name: str) -> Optional[Any]:
    log.msg('Checking for existing {}'.format(resource_type), name=resource_name)
    resource_list = list_resources()
    matching_resources = [d for d in resource_list.items
                         if d.metadata.name == resource_name]
    if len(matching_resources) > 2:
        raise ValueError('Found {} {}s matching `{}`'.format(len(matching_resources), resource_type, resource_name))
    if not matching_resources:
        return None
    return matching_resources[0]


def upsert_resource(existing_resource: Optional[Any], target_resource: Any, log: BoundLogger, resource_type: str, create: Callable[[], Any], patch: Callable[[], Any]) -> Any:
    if existing_resource is not None:
        if comparisons.specs_are_the_same(existing_resource, target_resource):
            log.msg('Existing {} does not need to be updated'.format(resource_type), name=existing_resource.metadata.name)
            new_resource = existing_resource
        else:
            log.msg('Updating existing {}'.format(resource_type))
            new_resource = patch()
            log.msg('Updated {}'.format(resource_type), name=new_resource.metadata.name)
    else:
        log.msg('Creating new {}'.format(resource_type))
        new_resource = create()
        log.msg('Created {}'.format(resource_type), name=new_resource.metadata.name)
    return new_resource
