import { describe, expect, mock, test } from "bun:test";
import { compileModule } from "svelte/compiler";
import { Doc } from "yjs";
import type { HocuspocusProvider } from "@hocuspocus/provider";
import type { IncidentCollaborationController } from "./collaboration.svelte";

// Bun does not compile runes. Compile the real controller and supply its external
// dependencies locally, avoiding module mocks that would affect other tests.
const source = await Bun.file(new URL("./collaboration.svelte.ts", import.meta.url)).text();
const javascript = new Bun.Transpiler({ loader: "ts" }).transformSync(source);
const compiled = compileModule(javascript, {
	filename: "collaboration.svelte.ts",
	generate: "client",
	dev: false,
}).js.code;
const controllerCode = compiled
	.slice(0, compiled.indexOf("const ctx ="))
	.replace(/^import .*;$/gm, "")
	.replace("export class", "class");
const runtime = await import(import.meta.resolve("svelte/internal/client"));

type Auth = { data: { name: string; serverUrl: string; token: string } };
type Request = { path: { id: string } };
type ProviderConfig = ConstructorParameters<typeof HocuspocusProvider>[0];

function setup(
	request: (request: Request) => Promise<Auth> = async ({ path }) => ({
		data: { name: path.id, serverUrl: "ws://localhost", token: "token" },
	})
) {
	const providers: FakeProvider[] = [];
	let retry!: (count: number, error: { status?: number }) => boolean;
	let unmount!: () => void;
	class FakeProvider {
		configuration = {
			websocketProvider: { disconnect: mock(() => {}), connect: mock(async () => {}) },
		};
		destroy = mock(() => {});
		constructor(public config: ProviderConfig) {
			providers.push(this);
		}
	}
	const Controller = new Function(
		"$",
		"HocuspocusProvider",
		"WebSocketStatus",
		"Doc",
		"watch",
		"onMount",
		"createMutation",
		"requestDocumentSessionAuthMutation",
		`${controllerCode}\nreturn IncidentCollaborationController;`
	)(
		runtime,
		FakeProvider,
		{ Disconnected: "disconnected" },
		Doc,
		() => {},
		(mount: () => () => void) => {
			unmount = mount();
		},
		(options: () => { retry: typeof retry }) => {
			retry = options().retry;
			return { mutateAsync: request };
		},
		() => ({})
	) as typeof IncidentCollaborationController;
	const controller = new Controller(() => undefined);
	return { controller, providers, retry, unmount: () => unmount() };
}

function deferredAuth() {
	let resolve!: (value: Auth) => void;
	const promise = new Promise<Auth>((done) => {
		resolve = done;
	});
	return { promise, resolve };
}

const auth = (name: string): Auth => ({ data: { name, serverUrl: "ws://localhost", token: name } });

describe("incident collaboration lifecycle", () => {
	test("retries only transient auth errors, with a two-retry limit", () => {
		const { retry, unmount } = setup();
		for (const status of [undefined, 0, 408, 429, 500, 503]) {
			expect(retry(0, { status })).toBe(true);
			expect(retry(1, { status })).toBe(true);
			expect(retry(2, { status })).toBe(false);
		}
		for (const status of [400, 401, 403, 404, 422]) expect(retry(0, { status })).toBe(false);
		unmount();
	});

	test("ignores an old auth response after A → B → A navigation", async () => {
		const first = deferredAuth();
		let calls = 0;
		const { controller, providers, unmount } = setup(async ({ path }) => {
			calls += 1;
			return calls === 1 ? first.promise : auth(path.id);
		});
		const oldRequest = controller.connect("A");
		await controller.connect("B");
		await controller.connect("A");
		first.resolve(auth("A"));
		await oldRequest;
		expect(providers).toHaveLength(2);
		expect(providers[0].destroy).toHaveBeenCalledTimes(1);
		expect(providers[1].destroy).not.toHaveBeenCalled();
		unmount();
	});

	test("cleans up unsynchronized providers and rejects stale callbacks", async () => {
		const { controller, providers, unmount } = setup();
		await controller.connect("A");
		const old = providers[0];
		const document = old.config.document!;
		await controller.connect("B");
		expect(old.configuration.websocketProvider.disconnect).toHaveBeenCalledTimes(1);
		expect(old.destroy).toHaveBeenCalledTimes(1);
		expect(document.isDestroyed).toBe(true);
		old.config.onAuthenticated?.({ scope: "read-write" });
		old.config.onSynced?.({ state: true });
		old.config.onAuthenticationFailed?.({ reason: "old error" });
		expect(controller.canEdit).toBe(false);
		expect(controller.initialSynced).toBe(false);
		expect(controller.error).toBeUndefined();
		unmount();
		expect(providers[1].destroy).toHaveBeenCalledTimes(1);
	});

	test("reconnects with fresh credentials while preserving document edits", async () => {
		let tokens = 0;
		const { controller, providers, unmount } = setup(async ({ path }) => ({
			data: { ...auth(path.id).data, token: String(++tokens) },
		}));
		await controller.connect("A");
		const provider = providers[0];
		const getToken = provider.config.token as () => Promise<string>;
		expect(await getToken()).toBe("1");
		const document = provider.config.document!;
		document.getText("summary").insert(0, "Unsynced draft");
		provider.config.onAuthenticated?.({ scope: "read-write" });
		expect(controller.canEdit).toBe(true);
		controller.retry();
		expect(await getToken()).toBe("2");
		expect(providers).toHaveLength(1);
		expect(document.getText("summary").toString()).toBe("Unsynced draft");
		expect(document.isDestroyed).toBe(false);
		expect(provider.configuration.websocketProvider.connect).toHaveBeenCalledTimes(1);
		provider.config.onAuthenticated?.({ scope: "readonly" });
		expect(controller.canEdit).toBe(false);
		unmount();
	});
});
