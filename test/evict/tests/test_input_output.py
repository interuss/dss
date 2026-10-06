import sys
from datetime import datetime, timedelta, UTC
import json
import time
import logging
from typing import Any, Callable

from evict_helper import EvictHelper
from query_helper import QueryHelper

EVICT_FILE = "/tmp/evict_test_input_output.json"


def check_input_output(
    logger: logging.Logger,
    eh: EvictHelper,
    name: str,
    key: str,
    evict_args: dict[str, Any],
    create: Callable[[datetime], dict[str, Any] | None],
    get_id: Callable[[dict[str, Any]], str],
    get: Callable[[str], dict[str, Any] | None],
):
    logger.debug(f"Evicting all {name} older than 1s so only test entities expire")
    eh.run_evict(delete=True, **evict_args)

    t = datetime.now(UTC) + timedelta(seconds=1)

    logger.debug(f"Creating 2 test {name}")
    ids: list[str] = []
    for _ in range(2):
        entity = create(t)
        if not entity:
            logger.error(f"❌ Unable to create {name}")
            sys.exit(1)
        ids.append(get_id(entity))

    logger.debug(f"Waiting 3s so the {name} expire")
    _ = sys.stdout.flush()
    time.sleep(3)

    logger.debug(f"Listing expired {name} to {EVICT_FILE}")
    eh.run_evict(output_file=EVICT_FILE, **evict_args)

    listed = json.loads(eh.read_file(EVICT_FILE))[key] or []
    if not set(ids) <= set(listed):
        logger.error(
            f"❌ Expected {name} {ids} to be listed in the output, got {listed}"
        )
        sys.exit(1)

    logger.debug(f"Evicting only {name} {ids[0]} using an input file")
    eh.write_file(EVICT_FILE, json.dumps({key: [ids[0]]}))
    # TTLs cannot be set together with an input file
    input_args = {k: v for k, v in evict_args.items() if not k.endswith("_ttl")}
    eh.run_evict(delete=True, input_file=EVICT_FILE, **input_args)

    if get(ids[0]):
        logger.error(f"❌ Test {name} {ids[0]} shall have been deleted by evict")
        sys.exit(1)
    if not get(ids[1]):
        logger.error(f"❌ Test {name} {ids[1]} shall not have been deleted by evict")
        sys.exit(1)

    logger.debug(f"Evicting the remaining {name}")
    eh.run_evict(delete=True, **evict_args)


def test_input_output(qh: QueryHelper, eh: EvictHelper):
    logger = logging.getLogger("test_input_output")

    logger.info("📋 Input/output test")

    check_input_output(
        logger,
        eh,
        "RID ISAs",
        "rid_isas",
        {"rid_isa": True, "rid_ttl": "1s"},
        qh.create_rid_ISA,
        lambda isa: str(isa["service_area"]["id"]),
        qh.get_rid_ISA,
    )

    check_input_output(
        logger,
        eh,
        "RID subscriptions",
        "rid_subscriptions",
        {"rid_sub": True, "rid_ttl": "1s"},
        qh.create_rid_subscription,
        lambda sub: str(sub["subscription"]["id"]),
        qh.get_rid_subscription,
    )

    # Operational intents are spread apart so they don't conflict with each other
    op_intent_lngs = iter(56 + 0.01 * i for i in range(2))
    check_input_output(
        logger,
        eh,
        "SCD operational intents",
        "operational_intents",
        {"scd_oir": True, "scd_ttl": "1s"},
        lambda until: qh.create_scd_op_intent(until, lng=next(op_intent_lngs)),
        lambda op_intent: str(op_intent["operational_intent_reference"]["id"]),
        qh.get_scd_op_intent,
    )

    check_input_output(
        logger,
        eh,
        "SCD subscriptions",
        "scd_subscriptions",
        {"scd_sub": True, "scd_ttl": "1s"},
        qh.create_scd_subscription,
        lambda sub: str(sub["subscription"]["id"]),
        qh.get_scd_subscription,
    )

    logger.info("✅ Input/output test successful :)")
