from monitoring.mock_uss import webapp


@webapp.route("/msgsigning/status")
def msgsigning_status():
    return "Mock Message Signing Service Provider ok"


from . import routes_msgsigning