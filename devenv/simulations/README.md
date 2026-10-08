# Simulations

Simulated services that stand in for a partner's system: real telemetry in `otel-lgtm`, and a real alert
through Alertmanager. Demos and launch-plan checks run against them. These are not the backend's agent
evaluation scenarios.

Three services (`checkout-api`, `payments-api`, `search-api`) send request metrics and logs to
`otel-lgtm` over OTLP. Setting it unhealthy slows `payments-api`, so about 30% of `checkout-api`
requests fail, and `CheckoutHighErrorRate` fires in Alertmanager within about two minutes.
`search-api` does not change.

```sh
just sim           # starts telemetry and Alertmanager, then runs the generator until Ctrl-C
just sim set-mode unhealthy
just sim set-mode healthy
```

Under Paseo, run the `sim` workspace script instead of a terminal. Alertmanager requires `SLACK_WEBHOOK_URL` in
`.env`; it sends the alert to Slack as well.

Telemetry and Alertmanager are shared by every worktree, so run one simulation at a time: two would send the
same services into one Prometheus, and the alert would flap.

The mode is kept in `devenv/simulations/.mode` (gitignored) and read every second, so the mode can be set from any terminal. Healthy is the default.

## Deployments

Setting the mode unhealthy first reports the deploy that precedes the failure: a `succeeded` deployment of
`checkout-api` to `development`, with a fresh ID, a version such as `checkout-api@<unix time>` and no
repository. To receive it in Rezible:

1. In settings, install the Webhook integration with the Deployments preset and generate its webhook URL.
2. Run `just sim deployments-target URL` with that URL. It is kept in `devenv/simulations/.deployments-url`
   (gitignored) and used as is: the host reaches the backend through its HTTPS alias.

Without the file, `just sim set-mode unhealthy` says how to set it and switches the mode anyway; a failed
report also leaves the switch to go ahead. The deployment lands on the same `checkout-api` service as the
alert that follows.

Where to look:

- Grafana at `http://localhost:$GRAFANA_PORT`, Explore: Loki `{service_name="checkout-api"}`
  (add `| detected_level="error"` for failures), and Prometheus
  `sum by (job) (rate(http_server_request_duration_seconds_count[1m]))`.
- Alertmanager at `http://localhost:$ALERTMANAGER_PORT`: `CheckoutHighErrorRate` with
  `service="checkout-api"`, `env="development"`, `severity="critical"`.

Files: `generator.ts` (the services), `prometheus-alerts.yaml` (the alert definition, mounted into
`otel-lgtm`; its Prometheus configuration is `../configs/prometheus.yaml`).
