from monitoring.uss_qualifier.rid.reports import Issue


class TestRunnerError(RuntimeError):
    """An error encountered when interacting with a DSS or a USS"""
    def __init__(self, msg, issue: Issue):
        super(TestRunnerError, self).__init__(msg)
        self.issue = issue


class TestStepError(RuntimeError):
    """An error encountered when interacting with a DSS or a USS"""
    def __init__(self, msg, step):
        super(TestStepError, self).__init__(msg)
        self.step = step

