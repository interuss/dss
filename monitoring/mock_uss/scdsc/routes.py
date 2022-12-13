from monitoring.mock_uss import webapp


@webapp.route("/scdsc/status")
def scdsc_status():
    return "Mock SCD strategic coordinator ok"


from . import routes_scdsc
from . import routes_injection
