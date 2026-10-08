import { describe, expect, mock, test } from "bun:test";
import { parseAbsoluteToLocal } from "@internationalized/date";
import { compileModule } from "svelte/compiler";

import type { ApiError, GetUserSessionResponseBody } from "$lib/api";

import type { UserSessionState } from "./user-session.svelte";

// Bun does not compile runes. Compile the real session state and supply its external dependencies locally,
// avoiding module mocks that would affect other tests.
const source = await Bun.file(new URL("./user-session.svelte.ts", import.meta.url)).text();
const javascript = new Bun.Transpiler({ loader: "ts" }).transformSync(source);
const compiled = compileModule(javascript, {
	filename: "user-session.svelte.ts",
	generate: "client",
	dev: false,
}).js.code;
const stateCode = compiled
	.slice(0, compiled.indexOf("const ctx ="))
	.replace(/^import .*;$/gm, "")
	.replace("export class", "class");
const runtime = await import(import.meta.resolve("svelte/internal/client"));

type SessionQuery = {
	data?: GetUserSessionResponseBody;
	error: ApiError | null;
	isFetched: boolean;
	refetch: () => Promise<void>;
};

function setup() {
	const query: SessionQuery = runtime.proxy({
		data: undefined,
		error: null,
		isFetched: false,
		refetch: mock(async () => {}),
	});
	let checkExpiry!: () => void;
	const page = { route: { id: "/situations" }, url: { pathname: "/situations" } };
	const watch = <T>(getter: () => T, callback: (value: T) => void) => {
		runtime.render_effect(() => {
			const value = getter();
			runtime.untrack(() => callback(value));
		});
	};

	const State = new Function(
		"$",
		"page",
		"navigating",
		"getUserSessionOptions",
		"parseAbsoluteToLocal",
		"createQuery",
		"watch",
		"onMount",
		"tick",
		"beforeNavigate",
		"goto",
		"resolve",
		"setInterval",
		`${stateCode}\nreturn UserSessionState;`
	)(
		runtime,
		page,
		{ to: null },
		() => ({}),
		parseAbsoluteToLocal,
		() => query,
		watch,
		(mount: () => void) => mount(),
		async () => {},
		() => {},
		mock(async () => {}),
		(path: string) => path,
		(callback: () => void) => {
			checkExpiry = callback;
		}
	) as new () => UserSessionState;

	let state!: UserSessionState;
	const destroy = runtime.effect_root(() => {
		state = new State();
	});
	runtime.flush();

	const load = (expiresAt: Date) => {
		query.data = {
			data: {
				user: { id: "user-1" },
				organization: { id: "org-1", attributes: { setupRequired: false } },
				organizationRole: "member",
				expiresAt: expiresAt.toISOString(),
			},
		} as GetUserSessionResponseBody;
		query.error = null;
		query.isFetched = true;
		runtime.flush();
	};
	const fail = (error: ApiError) => {
		query.error = error;
		runtime.flush();
	};

	return { state, query, load, fail, checkExpiry: () => checkExpiry(), destroy };
}

const inAnHour = () => new Date(Date.now() + 60 * 60 * 1000);
const unauthenticated: ApiError = { status: 401, code: "unauthenticated", detail: "Sign in to continue." };

describe("UserSessionState", () => {
	test("hadSession becomes true when a session loads and stays true after it ends", () => {
		const { state, load, fail, destroy } = setup();
		expect(state.hadSession).toBe(false);

		load(inAnHour());
		expect(state.isAuthenticated).toBe(true);
		expect(state.hadSession).toBe(true);

		fail(unauthenticated);
		expect(state.isAuthenticated).toBe(false);
		expect(state.error).toEqual(unauthenticated);
		expect(state.hadSession).toBe(true);
		destroy();
	});

	test("someone never signed in has no session to have had", () => {
		const { state, fail, destroy } = setup();

		fail(unauthenticated);

		expect(state.isAuthenticated).toBe(false);
		expect(state.hadSession).toBe(false);
		destroy();
	});

	test("an expired session is refetched", () => {
		const { query, load, checkExpiry, destroy } = setup();

		load(inAnHour());
		checkExpiry();
		expect(query.refetch).not.toHaveBeenCalled();

		load(new Date(Date.now() - 1000));
		checkExpiry();
		expect(query.refetch).toHaveBeenCalledTimes(1);
		destroy();
	});
});
