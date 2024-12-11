import { Database } from "@hocuspocus/extension-database";
import pg from 'pg';
import type { fetchPayload, storePayload } from "@hocuspocus/server";

const retrosTableName = "retrospective_documents";
const schema = `
	CREATE TABLE IF NOT EXISTS ${retrosTableName} (
		"name" VARCHAR(255) NOT NULL UNIQUE,
		"data" BYTEA NOT NULL
	)`;

const selectQuery = `SELECT data FROM ${retrosTableName} WHERE name = $1::TEXT LIMIT 1`
const upsertQuery = `
	INSERT INTO ${retrosTableName} ("name", "data") 
		VALUES ($1::TEXT, $2::BYTEA)
	ON CONFLICT(name) DO UPDATE
		SET data = EXCLUDED.data`

export const createDatabase = async (connectionString: string) => {
	const pool = new pg.Pool({connectionString});
	
	try {
		await pool.query(schema);
	} catch (err) {
		throw new Error(err + " while connecting to database");
	}

	const fetch = async ({ documentName }: fetchPayload) => {
		const res = await pool.query(selectQuery, [documentName]);
		if (res.rowCount === 0) return null;
		return new Uint8Array(res.rows[0].data);
	}

	const store = async ({ documentName, state }: storePayload) => {
		await pool.query(upsertQuery, [documentName, state]);
	}

	return new Database({fetch, store});
}