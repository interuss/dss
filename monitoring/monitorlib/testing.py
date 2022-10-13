from monitoring.monitorlib.formatting import make_datetime


def assert_datetimes_are_equal(t1, t2, tolerance_seconds: float = 0) -> None:
    try:
        t1_datetime = make_datetime(t1)
        t2_datetime = make_datetime(t2)
    except ValueError as e:
        assert False, "Error interpreting value as datetime: {}".format(e)
    if tolerance_seconds == 0:
        assert t1_datetime == t2_datetime
    else:
        assert abs((t1_datetime - t2_datetime).total_seconds()) < tolerance_seconds
