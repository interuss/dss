from typing import List, Tuple

import flask

from monitoring.monitorlib import scd
from monitoring.monitorlib.clients import scd as scd_client
from implicitdict import ImplicitDict
from monitoring.mock_uss import resources, webapp
from monitoring.mock_uss.scdsc.database import db


@webapp.route("/scdsc/status")
def scdsc_status():
    return "Mock SCD strategic coordinator ok"


from . import routes_scdsc
from . import routes_injection
