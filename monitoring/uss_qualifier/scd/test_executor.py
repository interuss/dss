from monitoring.uss_qualifier.scd.utils import SCDQualifierTestConfiguration
from monitoring.uss_qualifier.utils import is_url


def validate_configuration(test_configuration: SCDQualifierTestConfiguration):
    try:
        for injection_target in test_configuration.injection_targets:
            is_url(injection_target.injection_base_url)
    except ValueError:
        raise ValueError("A valid url for injection_target must be passed")

def run_scd_tests(locale: str, test_configuration: SCDQualifierTestConfiguration,
                  auth_spec: str):
    # TODO Replace with actual implementation
    if locale is 'che':
        print("[SCD] Running ASTM and U-Space tests")
    else:
        print("[SCD] Running ASTM tests")