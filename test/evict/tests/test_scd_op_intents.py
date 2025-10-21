import sys
from datetime import datetime, timedelta, UTC
import time
import logging

from evict_helper import EvictHelper
from query_helper import QueryHelper


def test_scd_op_intents(qh: QueryHelper, eh: EvictHelper):
    logger = logging.getLogger("test_scd_op_intents")

    logger.info("📋 Operational intents test")

    t = datetime.now(UTC) + timedelta(seconds=1)

    logger.debug("Creating test operational intent")
    op_intent = qh.create_scd_op_intent(t)

    if not op_intent:
        logger.error("❌ Unable to create operational intent")
        sys.exit(1)

    op_intent_id: str = str(op_intent["operational_intent_reference"]["id"])

    logger.debug("Check that operational intent exists")
    if not qh.get_scd_op_intent(op_intent_id):
        logger.error("❌ Unable to retrieve operational intent after creation")
        sys.exit(1)

    logger.debug("Evicting operational intents older than 1s")
    eh.evict_scd_operational_intents("1s", delete=True)

    logger.debug("Check that operational intent still exists")
    if not qh.get_scd_op_intent(op_intent_id):
        logger.error(
            "❌ Test operational intent shall still be present since not expired"
        )
        sys.exit(1)

    logger.debug("Waiting 3s so the operational intent expires")
    _ = sys.stdout.flush()
    time.sleep(3)

    logger.debug("Evicting operational intents older than 1s in dry mode")
    eh.evict_scd_operational_intents("1s", delete=False)

    logger.debug("Check that operational intent still exists")
    if not qh.get_scd_op_intent(op_intent_id):
        logger.error(
            "❌ Test operational intent shall still be present since delete was set to false"
        )
        sys.exit(1)

    logger.debug("Evicting subscriptions older than 1s")
    eh.evict_scd_subscriptions("1s", delete=True)

    logger.debug("Check that operation intent still exists")
    if not qh.get_scd_op_intent(op_intent_id):
        logger.error(
            "❌ Test operation intent shall still be present since we evicted subscriptions"
        )
        sys.exit(1)

    logger.debug("Evicting operational intents older than 1s")
    eh.evict_scd_operational_intents("1s", delete=True)

    logger.debug("Check that operation intent has been deleted")
    if qh.get_scd_op_intent(op_intent_id):
        logger.error("❌ Test operation intent shall has been deleted by evict")
        sys.exit(1)

    logger.info("✅ SCD Operational intents test successful :)")
