import argparse

from monitoring.monitorlib.locality import Locality
from monitoring.uss_qualifier.scd.simulator.test_definitions.astm_strategic_coordination.nominal_planning import (
    NominalPlanningTestDefinition,
)


def parseArgs() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Generate test definitions using the simulator"
    )

    parser.add_argument(
        "--locale",
        required=True,
        help="Locale of the test definitions to generate with the simulator",
    )

    # TODO: Add reference time

    return parser.parse_args()


def main() -> int:
    args = parseArgs()
    locale = Locality(args.locale.upper())

    print(f"Starting generation of test definitions for locale {locale.upper()}")
    print(
        f"- Are intersections with same priority allowed ? {'yes' if locale.allow_same_priority_intersections else 'no'}"
    )
    print(f"- Is U-Space applicable ? {'yes' if locale.is_uspace_applicable else 'no'}")

    if locale.allow_same_priority_intersections:
        raise NotImplemented(
            "The simulator do not support intersections with same priority yet."
        )

    NominalPlanningTestDefinition(locale).save()
    # TODO: Generate others astm-strategic-coordination test definitions

    if locale.is_uspace_applicable:
        # TODO: Generate u-space test definitions
        pass


if __name__ == "__main__":
    exit(main())
