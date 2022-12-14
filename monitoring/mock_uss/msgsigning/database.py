import json
from typing import Dict, Optional

from monitoring.monitorlib.multiprocessing import SynchronizedValue
from implicitdict import ImplicitDict

#TODO Use this database to dynamically alter which key pair to use for message signing activities.

class Database(ImplicitDict):
    """Simple in-memory pseudo-database tracking whether or not to use a valid key pair for message signing activities."""

    use_valid_keypair: bool = True

db = SynchronizedValue(
    Database(),
    decoder=lambda b: ImplicitDict.parse(json.loads(b.decode("utf-8")), Database),
)
