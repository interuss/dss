from implicitdict import ImplicitDict
from monitoring.deployment_manager.systems.dss.v1.configuration import V1DSS


class DSS(ImplicitDict):
    """Definition of a DSS instance.

    Only one of the fields below should be populated.
    """

    v1: V1DSS
