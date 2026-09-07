import { Hocuspocus } from "@hocuspocus/server";
import { Logger } from "@hocuspocus/extension-logger";
import crossws from "crossws/adapters/bun";

import { loadConfig } from "./config";
import { DocumentsServerExtension } from "./server";

const runServer = async () => {
  const cfg = loadConfig();

  const hocuspocus = new Hocuspocus({
    name: cfg.name,
    timeout: 30000,
    debounce: 1000,
    maxDebounce: 30000,
    quiet: false,
    extensions: [new Logger(), new DocumentsServerExtension(cfg)],
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

  Bun.serve({
    hostname: cfg.host,
    port: cfg.port,
    websocket: ws.websocket,
    fetch(request, server) {
      if (request.headers.get("upgrade") === "websocket") {
        return ws.handleUpgrade(request, server);
      }
      if (new URL(request.url).pathname === "/health") {
        return new Response(null, { status: 204 });
      }
      return new Response("Not Found", { status: 404 });
    },
  });

  console.log(`running server on ${cfg.host}:${cfg.port}`);
};

try {
  runServer();
} catch (e: unknown) {
  if (e instanceof Error) {
    console.error("Failed to create server: %s", e.message);
  } else {
    console.error("Failed to create server: %s", e);
  }
}
