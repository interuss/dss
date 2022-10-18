from implicitdict import ImplicitDict, StringBasedTimeDelta

from monitoring.uss_qualifier.resources.resource import Resource


class EvaluationConfiguration(ImplicitDict):
    min_polling_interval: StringBasedTimeDelta = StringBasedTimeDelta("5s")
    """Do not repeat system observations with intervals smaller than this."""

    max_propagation_latency: StringBasedTimeDelta = StringBasedTimeDelta("10s")
    """Allow up to this much time for data to propagate through the system."""

    min_query_diagonal: float = 100
    """Do not make queries with diagonals smaller than this many meters."""

    repeat_query_rect_period: int = 3
    """If set to a value above zero, reuse the most recent query rectangle/view every this many queries."""


class EvaluationConfigurationResource(Resource[EvaluationConfiguration]):
    configuration: EvaluationConfiguration

    def __init__(self, specification: EvaluationConfiguration):
        self.configuration = specification
