from uas_standards.interuss.automated_testing.flight_planning.v1.api import (
    StatusResponseStatus,
    StatusResponse,
)
from monitoring.mock_uss import webapp
from monitoring.monitorlib.geoawareness_automated_testing.api import (
    SCOPE_GEOAWARENESS_TEST,
)
from ..auth import requires_scope
from ...monitorlib import versioning


@webapp.route("/geoawareness/status")
@requires_scope([SCOPE_GEOAWARENESS_TEST])
def geoawareness_status():
    return StatusResponse(
        status=StatusResponseStatus.Ready, version=versioning.get_code_version()
    )


from . import routes_geoawareness
