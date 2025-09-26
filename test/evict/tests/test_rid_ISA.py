import sys
from datetime import datetime, timedelta, UTC
import time
import logging

from evict_helper import EvictHelper
from query_helper import QueryHelper


def test_rid_ISA(qh: QueryHelper, eh: EvictHelper):
    logger = logging.getLogger("test_rid_ISA")

    logger.info("📋 RID ISA test")

    t = datetime.now(UTC) + timedelta(seconds=1)

    logger.debug("Creating test ISA")
    sub = qh.create_rid_ISA(t)

    if not sub:
        logger.error("❌ Unable to create ISA")
        sys.exit(1)

    ISA_id: str = str(sub["service_area"]["id"])

    logger.debug("Check that ISA exists")
    if not qh.get_rid_ISA(ISA_id):
        logger.error("❌ Unable to retrieve ISA after creation")
        sys.exit(1)

    logger.debug("Evicting subscriptions older than 1s")
    eh.evict_rid_ISAs("1s", delete=True)

    logger.debug("Check that ISA still exists")
    if not qh.get_rid_ISA(ISA_id):
        logger.error("❌ Test ISA shall still be present since not expired")
        sys.exit(1)

    logger.debug("Waiting 3s so the ISA expires")
    _ = sys.stdout.flush()
    time.sleep(3)

    logger.debug("Evicting subscriptions older than 1s in dry mode")
    eh.evict_rid_ISAs("1s", delete=False)

    logger.debug("Check that ISA still exists")
    if not qh.get_rid_ISA(ISA_id):
        logger.error("❌ Test ISA shall still be present since delete was set to false")
        sys.exit(1)

    logger.debug("Evicting subscriptions older than 1s")
    eh.evict_rid_subscriptions("1s", delete=True)

    logger.debug("Check that ISA still exists")
    if not qh.get_rid_ISA(ISA_id):
        logger.error(
            "❌ Test ISA shall still be present since we evicted subscriptions"
        )
        sys.exit(1)

    logger.debug("Evicting subscriptions older than 1s on another locality")
    eh.evict_rid_ISAs("1s", delete=True, locality="somethingelse")
    if not qh.get_rid_ISA(ISA_id):
        logger.error(
            "❌ Test ISA shall still be present since we used another locality"
        )
        sys.exit(1)

    logger.debug("Evicting subscriptions older than 1s")
    eh.evict_rid_ISAs("1s", delete=True)

    logger.debug("Check that ISA has been deleted")
    if qh.get_rid_ISA(ISA_id):
        logger.error("❌ Test ISA shall has been deleted by evict")
        sys.exit(1)

    logger.info("✅ RID ISA test successful :)")
