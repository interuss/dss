ALTER TABLE scd_subscriptions
SET (
    ttl_expiration_expression = 'ends_at',
    ttl_job_cron = '@daily'
);
