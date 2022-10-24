import json
import os

import shapely.geometry
from shapely.geometry import Point, LineString

from monitoring.uss_qualifier.resources.netrid import (
    FlightDataResource,
    FlightDataStorageResource,
)
from monitoring.uss_qualifier.scenarios.scenario import TestScenario


class StoreFlightData(TestScenario):
    _flights_data: FlightDataResource
    _storage_config: FlightDataStorageResource

    def __init__(
        self,
        flights_data: FlightDataResource,
        storage_configuration: FlightDataStorageResource,
    ):
        super().__init__()
        self._flights_data = flights_data
        self._storage_config = storage_configuration

    def run(self):
        self.begin_test_scenario()
        self.record_note(
            "Flight count",
            f"{len(self._flights_data.flight_collection.flights)} flights",
        )
        cfg = self._storage_config.storage_configuration
        if "flight_record_collection_path" in cfg:
            self.record_note(
                "flight_record_collection_path",
                f"Storing FlightRecordCollection to {os.path.abspath(cfg.flight_record_collection_path)}",
            )
        if "geojson_tracks_path" in cfg:
            self.record_note(
                "geojson_tracks_path",
                f"Storing GeoJSON tracks to {os.path.abspath(cfg.geojson_tracks_path)}",
            )
        self.begin_test_case("Store flight data")
        self.begin_test_step("Store flight data")

        if (
            "flight_record_collection_path"
            in self._storage_config.storage_configuration
            and self._storage_config.storage_configuration.flight_record_collection_path
        ):
            with open(
                self._storage_config.storage_configuration.flight_record_collection_path,
                "w",
            ) as f:
                json.dump(self._flights_data.flight_collection, f, indent=2)

        if (
            "geojson_tracks_path" in self._storage_config.storage_configuration
            and self._storage_config.storage_configuration.geojson_tracks_path
        ):
            flight_point_current_index = {}
            num_flights = len(self._flights_data.flight_collection.flights)
            for i in range(num_flights):
                flight_point_current_index[i] = 0

            for track_id, flight in enumerate(
                self._flights_data.flight_collection.flights
            ):
                feature_collection = {"type": "FeatureCollection", "features": []}
                point_collection = []
                for state in flight.states:
                    p = Point(
                        (state.position.lng, state.position.lat, state.position.alt)
                    )
                    point_collection.append(p)

                line = LineString(point_collection)
                line_feature = {
                    "type": "Feature",
                    "properties": {},
                    "geometry": shapely.geometry.mapping(line),
                }
                feature_collection["features"].append(line_feature)
                path_file_name = "track_%s.geojson" % str(
                    track_id + 1
                )  # Avoid Zero based numbering

                tracks_file_path = os.path.join(
                    self._storage_config.storage_configuration.geojson_tracks_path,
                    path_file_name,
                )
                with open(tracks_file_path, "w") as f:
                    f.write(json.dumps(feature_collection))

        self.end_test_step()
        self.end_test_case()
        self.end_test_scenario()
