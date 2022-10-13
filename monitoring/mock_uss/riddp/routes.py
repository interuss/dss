from monitoring.mock_uss import webapp


@webapp.route("/riddp/status")
def riddp_status():
    return "Mock RID Display Provider ok"


from . import routes_observation
from . import routes_behavior
