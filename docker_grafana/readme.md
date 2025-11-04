docker compose up

open grafana site:
127.0.0.1:3000
account: admin/admin


sum(demo_request_duration_seconds_sum) by (method, event)
/
sum(demo_request_duration_seconds_count) by (method, event)


sum(rate(demo_request_total{}[1m]))by(method, event)