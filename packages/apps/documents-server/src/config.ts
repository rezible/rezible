export type Config = {
	name: string;
	host: string;
	port: number;
	dbUrl: string;
	sessionTokenSecretKey: Uint8Array;
}

const loadDbUrl = () => {
	let dbUrl = process.env.DB_URL ?? "";
	if (!!dbUrl) return dbUrl;

	const pgRole = process.env.POSTGRES__USER ?? "documents";
	const pgPassword = process.env.POSTGRES__PASSWORD ?? "";
	if (!pgPassword) {
		throw new Error("postgres password empty");
	}
	const pgHost = process.env.POSTGRES__HOST ?? "localhost";
	const pgPort = process.env.POSTGRES__PORT ?? 5432;
	const pgDatabase = process.env.POSTGRES__DATABASE ?? "rezible";
	const pgSslMode = process.env.POSTGRES__SSLMODE ?? "require";
	return `postgresql://${pgRole}:${pgPassword}@${pgHost}:${pgPort}/${pgDatabase}?sslmode=${pgSslMode}`
}

export const pasetoLocalKeyFromHex = (hex: string): Uint8Array => {
	if (!/^[0-9a-fA-F]{64}$/.test(hex)) {
		throw new Error("document session token key must be 64 hex characters");
	}
	const bytes = new Uint8Array(hex.length / 2);
	for (let i = 0; i < bytes.length; i += 1) {
		bytes[i] = Number.parseInt(hex.slice(i * 2, i * 2 + 2), 16);
	}
	const prefix = new TextEncoder().encode("k4.local.");
	const key = new Uint8Array(prefix.length + bytes.length);
	key.set(prefix);
	key.set(bytes, prefix.length);
	return key;
};

export const loadConfig = (): Config => {
	const name = process.env.NAME ?? "documents-server";

	const host = process.env.HOST ?? "0.0.0.0";
	let port = Number.parseInt(process.env.PORT ?? "7002", 10);
	if (port < 1024) port = 7003;

	const dbUrl = loadDbUrl();

	const sessionTokenSecretKey = pasetoLocalKeyFromHex(process.env.DOCUMENTS__SESSION_TOKEN_SECRET_HEX ?? "");

	return { name, host, port, dbUrl, sessionTokenSecretKey };
}
