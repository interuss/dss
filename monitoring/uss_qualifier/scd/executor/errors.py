
class TestStepError(RuntimeError):
    """An error encountered when interacting with a DSS or a USS"""
    def __init__(self, msg, step):
        super(TestStepError, self).__init__(msg)
        self.step = step

