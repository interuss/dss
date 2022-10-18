import os
from typing import Optional

from implicitdict import ImplicitDict

from monitoring.monitorlib.auth import make_auth_adapter
from monitoring.monitorlib import infrastructure
from monitoring.uss_qualifier.resources.resource import Resource


class AuthAdapterSpecification(ImplicitDict):
    """Specification for an AuthAdapter resource.

    Exactly one of the fields defined must be populated.
    """

    auth_spec: Optional[str]
    """Literal representation of auth spec.  WARNING: Specifying this directly may cause sensitive information to be included in reports and unprotected configuration files."""

    environment_variable_containing_auth_spec: Optional[str]
    """Name of environment variable containing the auth spec.  This is the preferred method of providing the auth spec."""


class AuthAdapter(Resource[AuthAdapterSpecification]):
    adapter: infrastructure.AuthAdapter

    def __init__(self, specification: AuthAdapterSpecification):
        if specification.environment_variable_containing_auth_spec:
            if (
                specification.environment_variable_containing_auth_spec
                not in os.environ
            ):
                raise ValueError(
                    "Environment variable {} could not be found".format(
                        specification.environment_variable_containing_auth_spec
                    )
                )
            spec = os.environ[specification.environment_variable_containing_auth_spec]
        elif specification.auth_spec:
            spec = specification.auth_spec
        else:
            raise ValueError("No auth spec was declared")
        self.adapter = make_auth_adapter(spec)
