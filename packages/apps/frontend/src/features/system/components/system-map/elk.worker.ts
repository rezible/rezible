// Vite bundles this entry as a Web Worker so ELK layout runs off the UI thread.
// Explicit worker loading also avoids ELK's default entry misdetecting Bun's environment.
import "elkjs/lib/elk-worker.js";
