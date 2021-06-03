from monitoring.monitorlib.typing import ImplicitDict


class FlightPassportClientDetails(ImplicitDict):
    """ A object to hold instance and client details of a passport installation """

    client_id: str
    client_secret: str
