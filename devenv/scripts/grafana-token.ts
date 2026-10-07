// Manages a Viewer service account in the shared otel-lgtm Grafana for this workspace's Grafana integration.
// The Grafana allows anonymous admin access, so these requests need no credentials.

for (const name of ["GRAFANA_PORT", "WORKSPACE_ID"]) {
  if (!process.env[name]) {
    throw new Error(`${name} is required: run "just dev setup-workspace", then reopen devbox shell`);
  }
}

const grafanaURL = `http://localhost:${process.env.GRAFANA_PORT}`;
// Worktrees share one Grafana, so each workspace gets its own account.
const accountName = `rezible-${process.env.WORKSPACE_ID}`;

async function grafana<T>(method: string, path: string, body?: unknown): Promise<T> {
  let resp: Response;
  try {
    resp = await fetch(`${grafanaURL}${path}`, {
      method,
      headers: { "Content-Type": "application/json" },
      body: body === undefined ? undefined : JSON.stringify(body),
    });
  } catch {
    throw new Error(`cannot reach Grafana at ${grafanaURL}: run "just dev infra-up telemetry"`);
  }
  if (!resp.ok) {
    throw new Error(`${method} ${path}: ${resp.status} ${await resp.text()}`);
  }
  return (await resp.json()) as T;
}

async function deleteAccounts() {
  const query = encodeURIComponent(accountName);
  const found = await grafana<{ serviceAccounts: { id: number; name: string }[] }>(
    "GET",
    `/api/serviceaccounts/search?query=${query}`
  );
  const accounts = found.serviceAccounts.filter((account) => account.name === accountName);
  for (const account of accounts) {
    await grafana("DELETE", `/api/serviceaccounts/${account.id}`);
  }
  return accounts.length;
}

// Grafana shows a token only once, so creating replaces any earlier account and its tokens.
async function createToken() {
  await deleteAccounts();
  const account = await grafana<{ id: number }>("POST", "/api/serviceaccounts", {
    name: accountName,
    role: "Viewer",
  });
  const token = await grafana<{ key: string }>("POST", `/api/serviceaccounts/${account.id}/tokens`, {
    name: accountName,
  });
  console.log(`Created Viewer service account ${accountName}. Install Grafana in Rezible with:`);
  console.log(`  URL:   ${grafanaURL}`);
  console.log(`  token: ${token.key}`);
  console.log(`Then set the data source UIDs: loki for logs, prometheus for metrics.`);
}

switch (process.argv[2]) {
  case "create":
    await createToken();
    break;
  case "delete": {
    const deleted = await deleteAccounts();
    console.log(deleted > 0 ? `Deleted service account ${accountName}.` : `No service account ${accountName}.`);
    break;
  }
  default:
    throw new Error("usage: grafana-token.ts <create|delete>");
}
