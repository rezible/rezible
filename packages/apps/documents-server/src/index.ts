import { createDocumentsServer } from "./app";
import { loadConfig } from "./config";

try {
  const documentsServer = createDocumentsServer(loadConfig());
  documentsServer.listen();
  const onSignal = () => {
    documentsServer.shutdown().then((code) => process.exit(code));
  };
  process.on("SIGTERM", onSignal);
  process.on("SIGINT", onSignal);
} catch (e: unknown) {
  if (e instanceof Error) {
    console.error("Failed to create server: %s", e.message);
  } else {
    console.error("Failed to create server: %s", e);
  }
  process.exit(1);
}
