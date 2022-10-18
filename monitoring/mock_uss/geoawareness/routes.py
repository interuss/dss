from monitoring.mock_uss import webapp
from monitoring.monitorlib.geoawareness_automated_testing.api import (
    StatusResponse,
    SCOPE_GEOAWARENESS_TEST,
    HarnessStatus,
)
from ..auth import requires_scope
from ...monitorlib import versioning


@webapp.route("/geoawareness/status")
@requires_scope([SCOPE_GEOAWARENESS_TEST])
def geoawareness_status():
    return StatusResponse(
        status=HarnessStatus.Ready, version=versioning.get_code_version()
    )
