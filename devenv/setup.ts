import { $, SQL } from "bun";
import { randomUUID } from "node:crypto";
import { join } from "node:path";

const workspaceEnvFile = Bun.file(join(process.cwd(), "..", ".env.workspace"));

const dbContainerName = "rezible-postgres";
const dbContainerPort = process.env.POSTGRES_PORT ?? "7010";
const dbContainerVolume = "rezible-postgres-data";

const dbAdminUser = "postgres";
const dbAdminPassword = "password";

const dbAppUser = "rez_app";
const dbAppPassword = "rez_app";

const execDatabaseContainer = async (cmd: string) => {
    await $`docker exec -it ${dbContainerName} psql -U postgres -c '${cmd}'`
}

const createDatabaseContainer = async () => {
    await $`docker create --name ${dbContainerName} -p ${`${dbContainerPort}:5432`} \
        -e POSTGRES_USER=${dbAdminUser} \
        -e POSTGRES_PASSWORD=${dbAdminPassword} \
        -v ${dbContainerVolume}:/var/lib/postgresql/data \
        postgres:17`;
};

const destroyDatabaseContainer = async () => {
    await $`docker stop ${dbContainerName} && docker rm -v ${dbContainerName} && docker volume rm ${dbContainerVolume}`.quiet().nothrow();
}

const waitForDatabaseReady = async (waitMs: number, maxAttempts: number) => {
    console.log("waiting for database to be ready...");
    for (let attempt = 0; attempt < maxAttempts; attempt++) {
        const res = await $`docker exec -it ${dbContainerName} pg_isready`.quiet().nothrow();
        if (res.exitCode === 0) return;
        // console.log(`[attempt ${attempt + 1}] waiting ${waitMs}ms for database to be ready...`)
        await Bun.sleep(waitMs);
    }
    throw new Error("failed to get container ready");
}

const ensureDatabaseContainer = async () => {
    const { exitCode } = await $`docker container inspect ${dbContainerName}`.nothrow().quiet();
    const create = exitCode > 0;
    if (create) await createDatabaseContainer();

    const running = (await $`docker container inspect -f {{.State.Running}} ${dbContainerName}`.text()).trim();
    if (running !== "true") await $`docker start ${dbContainerName}`.quiet();
    await waitForDatabaseReady(1000, 5);

    if (create) await execDatabaseContainer(`CREATE USER ${dbAppUser} WITH PASSWORD '${dbAppPassword}';`);
};

const setupDatabase = async (dbName: string) => {
    await ensureDatabaseContainer();

    await execDatabaseContainer(`CREATE DATABASE "${dbName}"`);
    console.log(`created database "${dbName}"`)

    const pg = new SQL(`postgres://${dbAdminUser}:${dbAdminPassword}@localhost:${dbContainerPort}/${dbName}`);
    const statements = [
        `GRANT CONNECT ON DATABASE "${dbName}" TO ${dbAppUser}`,
        `CREATE SCHEMA IF NOT EXISTS river`,
        `CREATE SCHEMA IF NOT EXISTS rezible`,
        `GRANT USAGE ON SCHEMA rezible TO ${dbAppUser}`,
        `GRANT USAGE ON SCHEMA river TO ${dbAppUser}`,
        `ALTER DEFAULT PRIVILEGES IN SCHEMA rezible GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${dbAppUser}`,
        `ALTER DEFAULT PRIVILEGES IN SCHEMA rezible GRANT USAGE, SELECT ON SEQUENCES TO ${dbAppUser}`,
        `ALTER DEFAULT PRIVILEGES IN SCHEMA river GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO ${dbAppUser}`,
        `ALTER DEFAULT PRIVILEGES IN SCHEMA river GRANT USAGE, SELECT ON SEQUENCES TO ${dbAppUser}`,
        `ALTER ROLE ${dbAppUser} SET search_path TO rezible, river`,
    ];
    for (const s of statements) {
        try {
            await pg.unsafe(`${s};`);
        } catch (e) {
            console.error(`setting up user/db:\n "${s}"\n\t`, e.message);
        }
    }
};

const generateBranchId = () => randomUUID().replace("-", "").slice(0, 6);

const getFreePortRange = async (range: number) => {
    for (let i = 0; i < 10; i++) {
        const basePort = 1700 + ((range + 1) * i);
        let occupied = false;
        for (let p = 0; p < range; p++) {
            const free = await $`bunx detect-port-alt ${basePort + p}`.text();
            if (free.includes("occupied")) {
                occupied = true;
                break;
            }
        }
        if (!occupied) return basePort;
    }
    throw new Error("unable to get port free")
}

const writeWorkspaceEnv = async (workspaceId: string, dbName: string) => {
    const workspaceSuffix = workspaceId === "main" ? "" : `-${workspaceId}`;
    const devDomain = process.env.DEV_DOMAIN ?? "dev.rezible.com";
    const appDomain = `app${workspaceSuffix}.${devDomain}`;
    const apiDomain = `api${workspaceSuffix}.${devDomain}`;
    const authDomain = `accounts${workspaceSuffix}.${devDomain}`;

    const portRangeBase = await getFreePortRange(5);

    const envLines = [
        `WORKSPACE_ID=${workspaceId}`,
        `POSTGRES_PORT=${dbContainerPort}`,
        `APP_PORT=${portRangeBase}`,
        `BACKEND_PORT=${portRangeBase + 1}`,
        `AUTH_PORT=${portRangeBase + 2}`,

        `APP_DOMAIN=${appDomain}`,
        `APP_URL=https://${appDomain}`,
        `API_DOMAIN=${apiDomain}`,
        `API_URL=https://${apiDomain}`,
        `AUTH_DOMAIN=${authDomain}`,
        `AUTH_URL=https://${authDomain}`,

        `POSTGRES__PORT=${dbContainerPort}`,
        `POSTGRES__DATABASE=${dbName}`,
        "POSTGRES__SSLMODE=disable",
        `POSTGRES__ROLE_ADMIN__NAME=${dbAdminUser}`,
        `POSTGRES__ROLE_ADMIN__PASSWORD=${dbAdminPassword}`,
        `POSTGRES__ROLE_APP__NAME=${dbAppUser}`,
        `POSTGRES__ROLE_APP__PASSWORD=${dbAppPassword}`,
    ];

    await workspaceEnvFile.write(`${envLines.join("\n")}\n`);
};

const setupWorkspace = async () => {
    const forceSetup = process.env.FORCE_SETUP === "true";
    if (await workspaceEnvFile.exists()) {
        if (!forceSetup) {
            console.error(`workspace env already exists at ${workspaceEnvFile.name}\nrerun with --force to replace it`);
            return;
        }
        await workspaceEnvFile.delete();
        await destroyDatabaseContainer();
    }

    const branchName = (await $`git branch --show-current`.text()).trim();
    const isMainBranch = branchName === "main";
    const workspaceId = isMainBranch ? "main" : generateBranchId();
    const dbNameSuffix = workspaceId === "" ? "" : `-${workspaceId}`;
    const dbName = `rezible${dbNameSuffix}`;

    await setupDatabase(dbName);
    await writeWorkspaceEnv(workspaceId, dbName);
};

setupWorkspace();
