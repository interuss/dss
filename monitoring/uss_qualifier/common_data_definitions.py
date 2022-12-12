from enum import Enum


class Severity(str, Enum):
    Critical = "Critical"
    """The system under test has a critical problem that justifies the discontinuation of testing.
    
    This kind of issue not only makes the current test scenario unable to
    succeed, but is likely to cause spurious failures in other separate test
    scenarios as well.  This may occur, for instance, if the system was left
    dirty which is likely to prevent subsequent test scenarios to run correctly.
    This kind of issue should be rare as test scenarios should generally be
    mostly independent of each other.
    """

    High = "High"
    """The system under test has a problem that prevents the current test scenario from continuing.
    
    Error interrupts a test scenario but likely doesn't impact other, separate
    test scenarios.  For instance, the test step necessary to enable later test
    steps in the test scenario did not complete successfully.
    """

    Medium = "Medium"
    """The system does not meet requirements, but the current test scenario can continue.
    
    Further test steps will likely still result in reasonable evaluations.
    """

    Low = "Low"
    """The system meets requirements but could be improved.
    
    Further test steps can be executed without impact.  A test run with only
    Low-Severity issues will be considered successful.
    """
