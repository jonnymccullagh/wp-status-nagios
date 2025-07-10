# WP-Status Nagios Plugin

This script monitors the status of a WordPress installation by querying an endpoint and checking various metrics.

## Required Parameters

- -H <url> URL of the endpoint to check (required)
- -P <password> Authorization password for header (optional)

## Threshold parameters (optional):

- -Z <value> Warn if core updates are above threshold (default: 0)
- -Y <value> Warn if plugin updates are above threshold (default: 0)
- -X <value> Warn if theme updates are above threshold (default: 0)
- -W <value> Warn if unapproved comments are above threshold (default: 0)
- -V <value> Warn if response time (ms) is above threshold (default: 0.0)
- -U <value> Warn if peak memory usage (MB) is above threshold (default: 6.0)

## Example Usage

```
check_wp_status -H https://www.wp-status.net/wp-json/wp-status/v1/wp-status  -P topsecret
```

This example would return the following:

```
WARNING Plugin updates exceed threshold Theme updates exceed threshold | plugin_update_count=2 theme_update_count=1 core_update_available=false theme_update_count=0 response_time_ms=0.000870 peak_script_memory_mb=5.640000 wp_version=6.8.1 php_version=7.4.33 db_query_count=10
```
