from . import webapp


@webapp.route('/status')
def status():
  return 'RID system mock for rid_qualifier is Ok', 200


from . import routes_injection
from . import routes_observation
