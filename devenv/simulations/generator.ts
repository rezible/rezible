// Simulated services reporting OTLP request metrics and logs to otel-lgtm.
// The mode ("healthy" or "unhealthy") is read from devenv/simulations/.mode every tick.

const endpoint = `http://localhost:${process.env.OTEL_HTTP_PORT || "4318"}`;
const modeFile = Bun.file(new URL(".mode", import.meta.url));
const tickMs = 1000;
const startNs = nowNs();

type Mode = "healthy" | "unhealthy";
type Severity = "INFO" | "WARN" | "ERROR";
type Request = { status: number; seconds: number; severity: Severity; message: string };

type Service = {
  name: string;
  method: string;
  route: string;
  perSecond: number;
  request: (mode: Mode) => Request;
};

const services: Service[] = [
  {
    name: "checkout-api",
    method: "POST",
    route: "/api/checkout",
    perSecond: 5,
    request: (mode) => {
      if (mode === "unhealthy" && Math.random() < 0.3) {
        return {
          status: 502,
          seconds: jitter(3.02, 0.02),
          severity: "ERROR",
          message: "payment authorization failed: upstream timeout after 3000ms (payments-api)",
        };
      }
      if (Math.random() < 0.005) {
        return { status: 500, seconds: jitter(0.05), severity: "ERROR", message: "order total mismatch: cart changed during checkout" };
      }
      return ok(mode === "unhealthy" ? jitter(2.6, 0.1) : jitter(0.12));
    },
  },
  {
    name: "payments-api",
    method: "POST",
    route: "/v1/authorizations",
    perSecond: 5,
    request: (mode) => {
      if (mode === "healthy") return ok(jitter(0.08));
      const seconds = jitter(3, 0.15);
      if (Math.random() < 0.4) {
        return { status: 200, seconds, severity: "WARN", message: "connection pool exhausted (max=20)" };
      }
      return ok(seconds);
    },
  },
  {
    name: "search-api",
    method: "GET",
    route: "/v1/search",
    perSecond: 10,
    request: () => ok(jitter(0.04)),
  },
];

function ok(seconds: number): Request {
  return { status: 200, seconds, severity: "INFO", message: "request completed" };
}

// OpenTelemetry's advisory boundaries for http.server.request.duration.
const bounds = [0.005, 0.01, 0.025, 0.05, 0.075, 0.1, 0.25, 0.5, 0.75, 1, 2.5, 5, 7.5, 10];
type Histogram = { service: Service; status: number; count: number; sum: number; buckets: number[] };
const histograms = new Map<string, Histogram>();

function record(service: Service, request: Request) {
  const key = `${service.name}:${request.status}`;
  let histogram = histograms.get(key);
  if (!histogram) {
    histogram = { service, status: request.status, count: 0, sum: 0, buckets: new Array(bounds.length + 1).fill(0) };
    histograms.set(key, histogram);
  }
  histogram.count++;
  histogram.sum += request.seconds;
  const bucket = bounds.findIndex((bound) => request.seconds <= bound);
  histogram.buckets[bucket === -1 ? bounds.length : bucket]++;
}

const severityNumbers: Record<Severity, number> = { INFO: 9, WARN: 13, ERROR: 17 };

async function tick() {
  const mode = await readMode();
  const tickStart = nowNs() - BigInt(tickMs) * 1_000_000n;
  const logs = new Map<string, unknown[]>();

  for (const service of services) {
    const records: unknown[] = [];
    for (let i = poisson(service.perSecond); i > 0; i--) {
      const request = service.request(mode);
      record(service, request);
      const time = tickStart + BigInt(Math.floor(Math.random() * tickMs * 1_000_000));
      records.push({
        timeUnixNano: time.toString(),
        severityNumber: severityNumbers[request.severity],
        severityText: request.severity,
        body: { stringValue: `${service.method} ${service.route} ${request.status} ${Math.round(request.seconds * 1000)}ms: ${request.message}` },
        attributes: [
          stringAttribute("http.request.method", service.method),
          stringAttribute("http.route", service.route),
          intAttribute("http.response.status_code", request.status),
          intAttribute("duration_ms", Math.round(request.seconds * 1000)),
        ],
      });
    }
    logs.set(service.name, records);
  }

  const time = nowNs().toString();
  await Promise.all([
    post("/v1/metrics", {
      resourceMetrics: services.map((service) => ({
        resource: resource(service),
        scopeMetrics: [{
          scope: { name: "rezible-sim" },
          metrics: [{
            name: "http.server.request.duration",
            unit: "s",
            description: "Duration of HTTP server requests.",
            histogram: {
              aggregationTemporality: 2,
              dataPoints: [...histograms.values()]
                .filter((histogram) => histogram.service === service)
                .map((histogram) => ({
                  startTimeUnixNano: startNs.toString(),
                  timeUnixNano: time,
                  count: histogram.count.toString(),
                  sum: histogram.sum,
                  bucketCounts: histogram.buckets.map(String),
                  explicitBounds: bounds,
                  attributes: [
                    stringAttribute("http.request.method", service.method),
                    stringAttribute("http.route", service.route),
                    intAttribute("http.response.status_code", histogram.status),
                  ],
                })),
            },
          }],
        }],
      })),
    }),
    post("/v1/logs", {
      resourceLogs: services.map((service) => ({
        resource: resource(service),
        scopeLogs: [{ scope: { name: "rezible-sim" }, logRecords: logs.get(service.name) }],
      })),
    }),
  ]);
}

let lastText: string | undefined;
async function readMode(): Promise<Mode> {
  const text = (await modeFile.exists()) ? (await modeFile.text()).trim() : "healthy";
  const mode: Mode = text === "unhealthy" ? "unhealthy" : "healthy";
  if (text !== lastText) {
    if (text !== mode) console.warn(`unknown sim mode "${text}", using healthy`);
    console.log(`${new Date().toISOString()} sim mode: ${mode}`);
  }
  lastText = text;
  return mode;
}

let failing = false;
async function post(path: string, body: unknown) {
  try {
    const response = await fetch(endpoint + path, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(body),
    });
    if (!response.ok) throw new Error(`${response.status} ${await response.text()}`);
    if (failing) console.log(`${new Date().toISOString()} sending to ${endpoint} again`);
    failing = false;
  } catch (error) {
    if (!failing) console.warn(`${new Date().toISOString()} cannot send to ${endpoint}${path} (is telemetry running?): ${error}`);
    failing = true;
  }
}

function resource(service: Service) {
  return {
    attributes: [
      stringAttribute("service.name", service.name),
      stringAttribute("deployment.environment", "development"),
    ],
  };
}

function stringAttribute(key: string, value: string) {
  return { key, value: { stringValue: value } };
}

function intAttribute(key: string, value: number) {
  return { key, value: { intValue: value.toString() } };
}

function nowNs() {
  return BigInt(Date.now()) * 1_000_000n;
}

// Multiplies by a normally distributed factor (relative standard deviation `spread`), floored at 10%.
function jitter(seconds: number, spread = 0.25) {
  const normal = Math.sqrt(-2 * Math.log(1 - Math.random())) * Math.cos(2 * Math.PI * Math.random());
  return seconds * Math.max(0.1, 1 + spread * normal);
}

function poisson(mean: number) {
  const limit = Math.exp(-mean);
  let count = 0;
  for (let product = Math.random(); product > limit; product *= Math.random()) count++;
  return count;
}

console.log(`sending simulated telemetry to ${endpoint}; stop with Ctrl-C`);
setInterval(tick, tickMs);
