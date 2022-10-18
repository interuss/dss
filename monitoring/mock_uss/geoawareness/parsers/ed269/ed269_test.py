import json
import os
from datetime import timedelta
from implicitdict import ImplicitDict
from monitoring.mock_uss.geoawareness.parsers.ed269 import ED269Schema, ED269TimeType


def test_sample():
    with open(
        os.path.join(os.path.dirname(__file__), "ed269_test_sample_dataset.json")
    ) as f:
        data = json.load(f)

    ED269Schema.from_dict(data)


def test_timetype():
    class MyTimedData(ImplicitDict):
        t1: ED269TimeType
        t2: ED269TimeType
        t3: ED269TimeType
        t4: ED269TimeType

    data = ImplicitDict.parse({
        "t1": "12:34:56.78Z",
        "t2": "12:34Z",
        "t3": "12:34:56.78-0100",
        "t4": "00:00:00.00+0100"
    }, MyTimedData)

    assert data["t1"].time.hour == 12
    assert data["t1"].time.minute == 34
    assert data["t1"].time.second == 56
    assert data["t1"].time.microsecond == 780000
    assert data["t1"].time.utcoffset() == timedelta(hours=0)
    assert str(data['t1']) == "12:34:56.78Z"

    assert data["t2"].time.hour == 12
    assert data["t2"].time.minute == 34
    assert data["t2"].time.second == 0
    assert data["t2"].time.microsecond == 0
    assert data["t2"].time.utcoffset() == timedelta(hours=0)
    assert str(data['t2']) == "12:34:00.00Z"

    assert data["t3"].time.hour == 12
    assert data["t3"].time.minute == 34
    assert data["t3"].time.second == 56
    assert data["t3"].time.microsecond == 780000
    assert data["t3"].time.utcoffset() == timedelta(hours=-1)
    assert str(data['t3']) == "12:34:56.78-0100"

    assert data["t4"].time.hour == 0
    assert data["t4"].time.minute == 0
    assert data["t4"].time.second == 0
    assert data["t4"].time.microsecond == 0
    assert data["t4"].time.utcoffset() == timedelta(hours=1)
    assert str(data['t4']) == "00:00:00.00+0100"

    assert data["t3"].time > data["t2"].time # t2 is an hour earlier than t3
