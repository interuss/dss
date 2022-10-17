from typing import List, Optional

from implicitdict import ImplicitDict, StringBasedDateTime, StringBasedTimeDelta

from monitoring.monitorlib.rid import RIDAircraftState, RIDFlightDetails


class FullFlightRecord(ImplicitDict):
    reference_time: StringBasedDateTime
    """The reference time of this flight (usually the time of first telemetry)"""

    states: List[RIDAircraftState]
    """All telemetry that will be/was received for this flight"""

    flight_details: RIDFlightDetails
    """Details of this flight, as would be reported at the ASTM /details endpoint"""

    aircraft_type: str
    """Type of aircraft, as per RIDFlight.aircraft_type"""


class FlightRecordCollection(ImplicitDict):
    flights: List[FullFlightRecord]


class FlightDataJSONFileConfiguration(ImplicitDict):
    path: str
    """Path to a file containing a JSON representation of an instance of FlightRecordCollection.  This may be a local file or a web URL."""


class AdjacentCircularFlightsSimulatorConfiguration(ImplicitDict):
    minx: float = 7.4735784530639648
    """Western edge of bounding box (degrees longitude)"""

    miny: float = 46.9746744128218410
    """Southern edge of bounding box (degrees latitude)"""

    maxx: float = 7.4786210060119620
    """Eastern edge of bounding box (degrees longitude)"""

    maxy: float = 46.9776318195799121
    """Northern edge of bounding box (degrees latitude)"""

    utm_zone: str = "32T"
    """UTM Zone string for the location, see https://en.wikipedia.org/wiki/Universal_Transverse_Mercator_coordinate_system to identify the zone for the location."""

    altitude_of_ground_level_wgs_84 = 570
    """Height of the geoid above the WGS84 ellipsoid (using EGM 96) for Bern, rom https://geographiclib.sourceforge.io/cgi-bin/GeoidEval?input=46%B056%26%238242%3B53%26%238243%3BN+7%B026%26%238242%3B51%26%238243%3BE&option=Submit"""


class FlightDataKMLFileConfiguration(ImplicitDict):
    kml_path: str
    """Path to a local file containing a KML describing a FlightRecordCollection."""


class FlightDataSpecification(ImplicitDict):
    flight_start_delay: StringBasedTimeDelta = StringBasedTimeDelta("15s")
    """Amount of time between starting the test and commencement of flights"""

    json_file_source: Optional[FlightDataJSONFileConfiguration] = None
    """When this field is populated, flight data will be loaded from a JSON file"""

    kml_file_source: Optional[FlightDataKMLFileConfiguration] = None
    """When this field is populated, flight data will be generated from a KML file"""

    adjacent_circular_flights_simulation_source: Optional[
        AdjacentCircularFlightsSimulatorConfiguration
    ] = None
    """When this field is populated, flight data will be simulated with the AdjacentCircularFlightsSimulator"""
