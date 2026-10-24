import { Hocuspocus } from "@hocuspocus/server";
import { Logger } from "@hocuspocus/extension-logger";
import crossws from "crossws/adapters/bun";
import type { SQL } from "bun";

import type { Config } from "./config";
import { DocumentsServerExtension } from "./extension";

// Deployments must allow a longer termination grace period than this.
export const SHUTDOWN_TIMEOUT_MS = 30_000;
const READY_TIMEOUT_MS = 2_000;
const SHUTDOWN_POLL_MS = 50;

export type DocumentsServerOptions = {
  db?: SQL;
  shutdownTimeoutMs?: number;
};

// Settles with the promise, or rejects with the signal's reason once it aborts.
const untilAborted = <T>(promise: Promise<T>, signal: AbortSignal): Promise<T> =>
  Promise.race([
    promise,
    new Promise<never>((_, reject) => {
      if (signal.aborted) reject(signal.reason);
      signal.addEventListener("abort", () => reject(signal.reason), { once: true });
    }),
  ]);

// Closes every connection and stores pending writes. Settles once each document
// has unloaded, or its last store failed with nothing pending: Hocuspocus keeps
// such documents in memory. Hocuspocus neither signals when flushPendingStores
// completes nor surfaces store errors, so this polls.
const flushDocuments = async (
  hocuspocus: Hocuspocus,
  failedStores: ReadonlySet<string>,
  signal: AbortSignal,
) => {
  hocuspocus.closeConnections();
  while (!signal.aborted) {
    hocuspocus.flushPendingStores();
    const settled = [...hocuspocus.documents.values()].every(
      (document) =>
        failedStores.has(document.name) && hocuspocus.shouldUnloadDocument(document),
    );
    if (settled) return;
    await Bun.sleep(SHUTDOWN_POLL_MS);
  }
};

export const createDocumentsServer = (
  cfg: Config,
  { db, shutdownTimeoutMs = SHUTDOWN_TIMEOUT_MS }: DocumentsServerOptions = {},
) => {
  const extension = new DocumentsServerExtension(cfg, db);
  const hocuspocus = new Hocuspocus({
    name: cfg.name,
    timeout: 30000,
    debounce: 1000,
    maxDebounce: 30000,
    quiet: false,
    extensions: [new Logger(), extension],
  });

  const ws = crossws({
    hooks: {
      open(peer) {
        const wsLike = {
          get readyState() {
            return peer.websocket.readyState ?? 3; // 3 = CLOSED
          },
          send(data: any) {
            peer.send(data);
          },
          close(code?: number, reason?: string) {
            peer.close(code, reason);
          },
        };
        (peer as any)._hocuspocus = hocuspocus.handleConnection(
          wsLike,
          peer.request,
        );
      },
      message(peer, message) {
        (peer as any)._hocuspocus?.handleMessage(message.uint8Array());
      },
      close(peer, { code, reason }) {
        (peer as any)._hocuspocus?.handleClose({ code, reason });
      },
      error(peer, error) {
        console.error("WebSocket error for peer:", peer.id);
        console.error(error);
      },
    },
  });

  let server: ReturnType<typeof Bun.serve> | undefined;
  let shuttingDown = false;
  let shutdownPromise: Promise<number> | undefined;

  const isReady = async () => {
    if (shuttingDown) return false;
    try {
      await untilAborted(extension.ping(), AbortSignal.timeout(READY_TIMEOUT_MS));
      return true;
    } catch (error) {
      console.error("readiness database ping failed:", error);
      return false;
    }
  };

  const allowedOrigins = new Set(cfg.allowedOrigins);
  const fetch = async (
    request: Request,
    bunServer: Parameters<typeof ws.handleUpgrade>[1],
  ): Promise<Response | undefined> => {
    const pathname = new URL(request.url).pathname;
    if (pathname === "/health") {
      return new Response(null, { status: 204 });
    }
    if (pathname === "/ready") {
      return new Response(null, { status: (await isReady()) ? 200 : 503 });
    }
    if (request.headers.get("upgrade") === "websocket") {
      if (shuttingDown) {
        return new Response("Service Unavailable", { status: 503 });
      }
      const origin = request.headers.get("origin");
      if (!origin || !allowedOrigins.has(origin)) {
        return new Response("Forbidden", { status: 403 });
      }
      return ws.handleUpgrade(request, bunServer);
    }
    return new Response("Not Found", { status: 404 });
  };

  const listen = () => {
    server = Bun.serve({
      hostname: cfg.host,
      port: cfg.port,
      websocket: ws.websocket,
      fetch,
    });
    console.log(`running server on ${cfg.host}:${cfg.port}`);
  };

  // Returns the names of documents whose writes could not be stored.
  const closeAll = async (gracePeriod: AbortSignal): Promise<string[]> => {
    await flushDocuments(hocuspocus, extension.failedStores, gracePeriod);
    const unflushed = [...hocuspocus.documents.keys()];
    // The extension's onDestroy closes the database.
    await hocuspocus.hooks("onDestroy", { instance: hocuspocus });
    await server?.stop();
    return unflushed;
  };

  // Resolves with the process exit code.
  const runShutdown = async (): Promise<number> => {
    shuttingDown = true;
    const gracePeriod = AbortSignal.timeout(shutdownTimeoutMs);
    try {
      const unflushed = await untilAborted(closeAll(gracePeriod), gracePeriod);
      if (unflushed.length === 0) return 0;
      console.error("shutdown failed to store documents: %s", unflushed.join(", "));
    } catch (error) {
      if (gracePeriod.aborted) {
        console.error(
          "shutdown grace period of %dms expired before pending document writes were stored",
          shutdownTimeoutMs,
        );
      } else {
        console.error("shutdown failed:", error);
      }
    }
    return 1;
  };

  const shutdown = (): Promise<number> => {
    shutdownPromise ??= runShutdown();
    return shutdownPromise;
  };

  return { hocuspocus, fetch, listen, shutdown };
};
