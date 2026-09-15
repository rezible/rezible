import { SQL } from "bun";

for (const name of ["POSTGRES_APP_DB", "POSTGRES_ADMIN_USER", "POSTGRES_APP_USER", "POSTGRES_PORT"]) {
  if (!process.env[name]) {
    throw new Error(`${name} is required: run "just dev setup-workspace", then reopen devbox shell`);
  }
}

const database = process.env.POSTGRES_APP_DB!;
const adminUser = process.env.POSTGRES_ADMIN_USER!;
const appUser = process.env.POSTGRES_APP_USER!;
const port = process.env.POSTGRES_PORT!;
const adminURL = `postgres://${adminUser}:${adminUser}@localhost:${port}`;

async function verifyDatabase() {
  const pg = new SQL(`postgres://${appUser}:${appUser}@localhost:${port}/${database}`);
  try {
    await pg.unsafe("SELECT 1");
    console.log(`database ${database} is ready`);
  } finally {
    await pg.close();
  }
}

async function dropDatabase() {
  const pg = new SQL(`${adminURL}/postgres`);
  try {
    await pg.unsafe(`DROP DATABASE IF EXISTS "${database}" WITH (FORCE)`);
  } finally {
    await pg.close();
  }
}

async function ensureDatabase() {
  const pg = new SQL(`${adminURL}/postgres`);
  try {
    const roles = await pg`SELECT 1 FROM pg_roles WHERE rolname = ${appUser}`;
    if (roles.length === 0) {
      await pg.unsafe(`CREATE USER ${appUser} WITH PASSWORD '${appUser}'`);
    }
    const databases = await pg`SELECT 1 FROM pg_database WHERE datname = ${database}`;
    if (databases.length === 0) {
      await pg.unsafe(`CREATE DATABASE "${database}"`);
      console.log(`created database ${database}`);
    }

    const workspace = new SQL(`${adminURL}/${database}`);
    try {
      await workspace.unsafe(`GRANT CONNECT ON DATABASE "${database}" TO ${appUser}`);
      for (const schema of ["rezible", "river"]) {
        await workspace.unsafe(`CREATE SCHEMA IF NOT EXISTS ${schema}`);
        await workspace.unsafe(`GRANT USAGE ON SCHEMA ${schema} TO ${appUser}`);
        await workspace.unsafe(`ALTER DEFAULT PRIVILEGES IN SCHEMA ${schema} GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${appUser}`);
        await workspace.unsafe(`ALTER DEFAULT PRIVILEGES IN SCHEMA ${schema} GRANT USAGE, SELECT ON SEQUENCES TO ${appUser}`);
      }
      await workspace.unsafe(`ALTER ROLE ${appUser} SET search_path TO rezible, river`);
    } finally {
      await workspace.close();
    }
  } finally {
    await pg.close();
  }
}

switch (process.argv[2]) {
  case "ensure":
    await ensureDatabase();
    break;
  case "verify":
    await verifyDatabase();
    break;
  case "drop":
    await dropDatabase();
    break;
  default:
    throw new Error("usage: database.ts <ensure|verify|drop>");
}
