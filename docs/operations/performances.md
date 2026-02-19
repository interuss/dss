# Performances

## The SCD global lock option

!!! danger
     All DSS instances in a DSS pool must use the same value for this option. Mixing will result in dramatically lower performance.

     You can use the `/aux/v1/configuration/scd_lock_mode` endpoint to retrive the current value for a specifc DSS instance.

It has been reported in issue [#1311](https://github.com/interuss/dss/issues/1311) that creating a lot of overlapping operational intents may increase the datastore load in a way that creates timeouts.

By default, the code will try to lock on required subscriptions when working on operational intents, and having too many of them may lead to issues.

A solution to that is to switch to a global lock, that is just globally locking operational intents operations, regardless of subscriptions.

This will result in lower general throughput for operational intents that don't overlap, as only one of them can be processed at a time, but better performance in the issue's case as lock acquisition is simpler.

You should enable this option depending on your DSS usage/use case and what you want to maximize:
* If you have non-overlapping traffic and maximum global throughput, don't enable this flag
* If you have overlapping traffic and don't need high global throughput, enable this flag

The following graphs show example throughput without (on the left) and with the flag (on the right). This has been run on a local machine; on a real deployment you can expect lower performance (due to various latency), but similar relative numbers.

All graphs have been generated with the [loadtest present in the monitoring repository](https://github.com/interuss/monitoring/blob/main/monitoring/loadtest/README.md) using `SCD.py`.

![](../assets/perfs_scd_lock_overlapping.png)
*Overlapping requests. Notice the huge spikes on the left, as the datastore struggles to acquire locks.*

![](../assets/perfs_scd_lock_notoverlapping.png)
*Non-overlapping requests. Notice the reduction of performance on the right, with a single lock.*
