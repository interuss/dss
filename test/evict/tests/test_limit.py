import sys
from datetime import datetime, timedelta, UTC
import time
import logging
from typing import Any, Callable

from evict_helper import EvictHelper
from query_helper import QueryHelper

ENTITY_COUNT = 3
LIMIT = 2


def check_limit(
    logger: logging.Logger,
    name: str,
    create: Callable[[datetime], dict[str, Any] | None],
    get_id: Callable[[dict[str, Any]], str],
    get: Callable[[str], dict[str, Any] | None],
    evict: Callable[..., None],
):
    logger.debug(f"Evicting all {name} older than 1s so only test entities expire")
    evict("1s", delete=True)

    t = datetime.now(UTC) + timedelta(seconds=1)

    logger.debug(f"Creating {ENTITY_COUNT} test {name}")
    ids: list[str] = []
    for _ in range(ENTITY_COUNT):
        entity = create(t)
        if not entity:
            logger.error(f"❌ Unable to create {name}")
            sys.exit(1)
        ids.append(get_id(entity))

    logger.debug(f"Waiting 3s so the {name} expire")
    _ = sys.stdout.flush()
    time.sleep(3)

    logger.debug(f"Evicting {name} older than 1s with a limit of {LIMIT}")
    evict("1s", delete=True, limit=LIMIT)

    remaining = [i for i in ids if get(i)]
    if len(remaining) != ENTITY_COUNT - LIMIT:
        logger.error(
            f"❌ Expected {ENTITY_COUNT - LIMIT} {name} to remain after evicting with a limit of {LIMIT}, got {remaining}"
        )
        sys.exit(1)

    logger.debug(f"Evicting {name} older than 1s without limit")
    evict("1s", delete=True, limit=0)

    remaining = [i for i in ids if get(i)]
    if remaining:
        logger.error(f"❌ Test {name} {remaining} shall have been deleted by evict")
        sys.exit(1)


def test_limit(qh: QueryHelper, eh: EvictHelper):
    logger = logging.getLogger("test_limit")

    logger.info("📋 Limit test")

    check_limit(
        logger,
        "RID ISAs",
        qh.create_rid_ISA,
        lambda isa: str(isa["service_area"]["id"]),
        qh.get_rid_ISA,
        eh.evict_rid_ISAs,
    )

    check_limit(
        logger,
        "RID subscriptions",
        qh.create_rid_subscription,
        lambda sub: str(sub["subscription"]["id"]),
        qh.get_rid_subscription,
        eh.evict_rid_subscriptions,
    )

    # Operational intents are spread apart so they don't conflict with each other
    op_intent_lngs = iter(56 + 0.01 * i for i in range(ENTITY_COUNT))
    check_limit(
        logger,
        "SCD operational intents",
        lambda until: qh.create_scd_op_intent(until, lng=next(op_intent_lngs)),
        lambda op_intent: str(op_intent["operational_intent_reference"]["id"]),
        qh.get_scd_op_intent,
        eh.evict_scd_operational_intents,
    )

    check_limit(
        logger,
        "SCD subscriptions",
        qh.create_scd_subscription,
        lambda sub: str(sub["subscription"]["id"]),
        qh.get_scd_subscription,
        eh.evict_scd_subscriptions,
    )

    logger.info("✅ Limit test successful :)")
