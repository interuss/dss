from .resource import Resource, ResourceCollection


# TODO: Use this type annotation more widely
ResourceID = str
"""This plain string represents the ID/name of a resource"""

# TODO: Use this type annotation more widely
ResourceType = str
"""This plain string represents a type of resource, expressed as a Python class name qualified relative to this `resources` module"""
