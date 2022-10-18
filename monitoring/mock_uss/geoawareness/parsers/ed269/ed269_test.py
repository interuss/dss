import json
import os
from monitoring.mock_uss.geoawareness.parsers.ed269 import ED269Schema


def test_sample():
    with open(
        os.path.join(os.path.dirname(__file__), "ed269_test_sample_dataset.json")
    ) as f:
        data = json.load(f)

    ED269Schema.from_dict(data)
