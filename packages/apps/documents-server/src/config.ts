export type Config = {
	name: string;
	host: string;
	port: number;
	apiUrl: string;
	dbUrl: string;
}

const loadDbUrl = () => {
	let dbUrl = process.env.DB_URL ?? "";
	if (!!dbUrl) return dbUrl;

	const pgRole = process.env.POSTGRES__ROLE_DOCUMENTS__NAME ?? "documents";
	const pgPassword = process.env.POSTGRES__ROLE_DOCUMENTS__PASSWORD ?? "";
	if (!pgPassword) {
		throw new Error("postgres password empty");
	}
	const pgHost = process.env.POSTGRES__HOST ?? "localhost";
	const pgPort = process.env.POSTGRES__PORT ?? 5432;
	const pgDatabase = process.env.POSTGRES__DATABASE ?? "rezible";
	const pgSslMode = process.env.POSTGRES__SSLMODE ?? "require";
	return `postgresql://${pgRole}:${pgPassword}@${pgHost}:${pgPort}/${pgDatabase}?sslmode=${pgSslMode}`
}

export const loadConfig = (): Config => {
	const name = process.env.NAME ?? "documents-server";

	const host = process.env.HOST ?? "0.0.0.0";
	let port = Number.parseInt(process.env.PORT ?? "7002", 10);
	if (port < 1024) port = 7003;

	const apiUrl = process.env.API_URL ?? "";
	const dbUrl = loadDbUrl();

	return { name, host, port, apiUrl, dbUrl };
}