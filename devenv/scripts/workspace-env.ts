import { randomUUID } from "node:crypto";

const run = (args: string[]) => {
  const result = Bun.spawnSync(args);
  if (result.exitCode !== 0) throw new Error(result.stderr.toString());
  return result.stdout.toString().trim();
}

const getWorkspaceId = (branch: string) => {
    const root = run(["git", "rev-parse", "--show-toplevel"]);
    const slug = (branch || "detached")
        .toLowerCase()
        .replace(/[^a-z0-9]+/g, "-")
        .replace(/^-|-$/g, "")
        .slice(0, 28);
    const hash = new Bun.CryptoHasher("sha256").update(root).digest("hex").slice(0, 8);
    return `${slug}-${hash}`;
};

const getDocumentSessionKeys = () => {
  const keysOutput = run(["bun", "run", "--silent", "--cwd=packages/apps/documents-server", "generate-session-keys"]);
  const { seedHex, publicKeyHex } = JSON.parse(keysOutput);
  return {
    "DOCUMENTS__SESSION_SIGNING_SEED_HEX": seedHex as string,
    "DOCUMENTS__SESSION_PUBLIC_KEY_HEX": publicKeyHex as string,
  };
};

const generateWorkspaceEnv = async () => {
  const workspaceFile = Bun.file(new URL("../.workspace-env.sh", import.meta.url));
  const previous: Record<string, string> = {};
  if (await workspaceFile.exists()) {
    const contents = await workspaceFile.text();
    for (const name of ["WORKSPACE_ID", "POSTGRES_APP_DB", "OIDC_CLIENT_ID", "OIDC_CLIENT_SECRET"]) {
      const match = contents.match(new RegExp(`^export ${name}='([^']+)'$`, "m"));
      if (!match) throw new Error(`Missing ${name} in workspace environment`);
      previous[name] = match[1];
    }
  }

  const devDomain = "dev.rezible.com";
  const authUrl = `https://auth.${devDomain}`;
  const postgresPort = "7010";
  const postgresAdminUser = "postgres";
  const postgresAppUser = "rez_app";
  const branch = run(["git", "branch", "--show-current"]);
  const workspaceId = previous.WORKSPACE_ID ?? getWorkspaceId(branch);
  const domainPrefix = branch === "main" ? "" : `${workspaceId}.`;

  const appDomain = `${domainPrefix}app.${devDomain}`;
  const apiDomain = `${domainPrefix}api.${devDomain}`;
  const documentsDomain = `${domainPrefix}documents.${devDomain}`;

  const database = previous.POSTGRES_APP_DB ?? `rezible-${workspaceId}`;
  const clientId = previous.OIDC_CLIENT_ID ?? `rezible-${workspaceId}`;
  const clientSecret = previous.OIDC_CLIENT_SECRET ?? randomUUID().replaceAll("-", "");

  const values = {
    DEV_DOMAIN: devDomain,
    POSTGRES_PORT: postgresPort,
    POSTGRES_TEST_PORT: "7011",
    DEX_PORT: "7012",
    DEX_GRPC_PORT: "7013",
    GRAFANA_PORT: "7014",
    ALERTMANAGER_PORT: "7015",
    POSTGRES_ADMIN_USER: postgresAdminUser,
    POSTGRES_APP_USER: postgresAppUser,
    POSTGRES_TEST_HOST: "localhost",
    POSTGRES_TEST_DB: "postgres",
    OTEL_TRACES_EXPORTER: "otlp",
    OTEL_METRICS_EXPORTER: "otlp",
    OTEL_LOGS_EXPORTER: "none",
    OTEL_EXPORTER_OTLP_ENDPOINT: "http://localhost:4317",
    OTEL_EXPORTER_OTLP_PROTOCOL: "grpc",
    OTEL_SERVICE_NAME: "rezible-backend",
    APP__DEBUG_MODE: "true",
    APP__SINGLETENANT__ENABLED: "true",
    DOCUMENTS__ALLOWED_ORIGINS: `https://${appDomain}`,
    HTTP__BASE_PATH: "",
    HTTP__AUTH__SESSION_SECRET: "superduuuuuuuuuuuuuuuuupersecret",
    POSTGRES__HOST: "localhost",
    POSTGRES__SSLMODE: "disable",
    WORKSPACE_ID: workspaceId,
    AUTH_URL: authUrl,
    POSTGRES_APP_DB: database,
    APP_URL: `https://${appDomain}`,
    API_URL: `https://${apiDomain}`,
    DOCUMENTS_URL: `https://${documentsDomain}`,
    OIDC_CLIENT_ID: clientId,
    OIDC_CLIENT_SECRET: clientSecret,
    APP__FRONTEND_DOMAIN: appDomain,
    APP__API_DOMAIN: apiDomain,
    DOCUMENTS__SERVER_URL: `https://${documentsDomain}`,
    HTTP__AUTH__OIDC__ISSUER: authUrl,
    HTTP__AUTH__OIDC__CLIENT_ID: clientId,
    HTTP__AUTH__OIDC__CLIENT_SECRET: clientSecret,
    POSTGRES__DATABASE: database,
    POSTGRES__PORT: postgresPort,
    POSTGRES__ROLE_ADMIN__NAME: postgresAdminUser,
    POSTGRES__ROLE_ADMIN__PASSWORD: postgresAdminUser,
    POSTGRES__ROLE_APP__NAME: postgresAppUser,
    POSTGRES__ROLE_APP__PASSWORD: postgresAppUser,
    ...getDocumentSessionKeys(),
  };

  const fileContents = Object.entries(values)
    .map(([name, value]) => `export ${name}='${value}'`)
    .join("\n") + "\n";
  await Bun.write(workspaceFile, fileContents);
}

await generateWorkspaceEnv();
