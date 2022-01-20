from monitoring.monitorlib import rid


VERTICES = [
    {
        'lng': 130.6205,
        'lat': -23.6558
    },
    {
        'lng': 130.6301,
        'lat': -23.6898
    },
    {
        'lng': 130.6700,
        'lat': -23.6709
    },
    {
        'lng': 130.6466,
        'lat': -23.6407
    },
]

CLOSER_POLYGON_VERTICES = [
    {
        'lng': 130.6205,
        'lat': -23.6558
    },
    {
        'lng': 130.6201,
        'lat': -23.6298
    },
    {
        'lng': 130.6500,
        'lat': -23.6209
    },
    {
        'lng': 130.6466,
        'lat': -23.6407
    },
]

GEO_POLYGON_STRING = rid.geo_polygon_string(VERTICES)

HUGE_VERTICES = [
    {
        'lng': 130,
        'lat': -23
    },
    {
        'lng': 130,
        'lat': -24
    },
    {
        'lng': 132,
        'lat': -24
    },
    {
        'lng': 132,
        'lat': -23
    },
]

HUGE_GEO_POLYGON_STRING = rid.geo_polygon_string(HUGE_VERTICES)

LOOP_VERTICES = [
    {
        'lat': -80.75088500976562,
        'lng': 37.045312802603355
    },
    {
        'lat': -80.2496337890625,
        'lng': 37.045312802603355
    },
    {
        'lat': -80.2496337890625,
        'lng': 37.35487607348372
    },
    {
        'lat': -80.75088500976562,
        'lng': 37.35487607348372
    },
    {
        'lat': -80.75088500976562,
        'lng': 37.045312802603355
    }
]

LOOP_GEO_POLYGON_STRING = rid.geo_polygon_string(LOOP_VERTICES)
