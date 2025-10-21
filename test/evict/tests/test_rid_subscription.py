import sys
from datetime import datetime, timedelta, UTC
import time
import logging

from evict_helper import EvictHelper
from query_helper import QueryHelper


def test_rid_subscription(qh: QueryHelper, eh: EvictHelper):
    logger = logging.getLogger("test_rid_subscription")

    logger.info("📋 RID Subscriptions test")

    t = datetime.now(UTC) + timedelta(seconds=1)

    logger.debug("Creating test subscription")
    sub = qh.create_rid_subscription(t)

    if not sub:
        logger.error("❌ Unable to create subscription")
        sys.exit(1)

    sub_id: str = str(sub["subscription"]["id"])

    logger.debug("Check that subscription exists")
    if not qh.get_rid_subscription(sub_id):
        logger.error("❌ Unable to retrieve subscription after creation")
        sys.exit(1)

    logger.debug("Evicting subscriptions older than 1s")
    eh.evict_rid_subscriptions("1s", delete=True)

    logger.debug("Check that subscription still exists")
    if not qh.get_rid_subscription(sub_id):
        logger.error("❌ Test subscription shall still be present since not expired")
        sys.exit(1)

    logger.debug("Waiting 3s so the subscription expires")
    _ = sys.stdout.flush()
    time.sleep(3)

    logger.debug("Evicting subscriptions older than 1s in dry mode")
    eh.evict_rid_subscriptions("1s", delete=False)

    logger.debug("Check that subscription still exists")
    if not qh.get_rid_subscription(sub_id):
        logger.error(
            "❌ Test subscription shall still be present since delete was set to false"
        )
        sys.exit(1)

    logger.debug("Evicting ISAs older than 1s")
    eh.evict_rid_ISAs("1s", delete=True)

    logger.debug("Check that subscription still exists")
    if not qh.get_rid_subscription(sub_id):
        logger.error("❌ Test subscription shall still be present since we evicted ISA")
        sys.exit(1)

    logger.debug("Evicting subscriptions older than 1s on another locality")
    eh.evict_rid_subscriptions("1s", delete=True, locality="somethingelse")
    if not qh.get_rid_subscription(sub_id):
        logger.error(
            "❌ Test subscription shall still be present since we used another locality"
        )
        sys.exit(1)

    logger.debug("Evicting subscriptions older than 1s")
    eh.evict_rid_subscriptions("1s", delete=True)

    logger.debug("Check that subscription has been deleted")
    if qh.get_rid_subscription(sub_id):
        logger.error("❌ Test subscription shall has been deleted by evict")
        sys.exit(1)

    logger.info("✅ RID Subscription test successful :)")
