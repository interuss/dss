from monitoring.rid_qualifier import create_flight_record_from_kml as frk

def test_get_interpolated_altitude(mocker):
    mocker.patch(
        'monitoring.rid_qualifier.create_flight_record_from_kml.get_polygons_distances_from_point',
        return_value=[1.1, 2.0, 3.3, 0.0]) 
    result_alt = frk.get_interpolated_altitude('point', 'polygons', [10,20,30, 40])
    assert result_alt == 40

    mocker.patch(
        'monitoring.rid_qualifier.create_flight_record_from_kml.get_polygons_distances_from_point',
        return_value=[10, 50, 100])
    result_alt = frk.get_interpolated_altitude('point', 'polygons', [10,20,30])
    assert abs(result_alt - 13.1) < 0.1

    mocker.patch(
        'monitoring.rid_qualifier.create_flight_record_from_kml.get_polygons_distances_from_point',
        return_value=[10, 50, 100])
    result_alt = frk.get_interpolated_altitude('point', 'polygons', [140,125,116])
    assert abs(result_alt - 135.84) < 0.1
    
