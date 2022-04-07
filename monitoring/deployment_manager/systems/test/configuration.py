from typing import Optional
from monitoring.monitorlib.typing import ImplicitDict


class TestV1(ImplicitDict):
    namespace: str = 'test'


class Test(ImplicitDict):
    v1: Optional[TestV1]
