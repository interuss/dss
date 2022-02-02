import flask
from werkzeug.exceptions import HTTPException

from monitoring.monitorlib import versioning, auth_validation
from monitoring.mock_uss import webapp


@webapp.route('/riddp/status')
def riddp_status():
    return 'Mock RID Display Provider ok'


from . import routes_observation
