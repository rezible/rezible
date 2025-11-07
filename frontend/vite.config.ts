import { sveltekit } from "@sveltejs/kit/vite";
import { defineConfig } from "vite";

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		host: "127.0.0.1",
		port: 5173,
		strictPort: true,
		allowedHosts: ["app.rezible.test"],
		proxy: {
			"/api": {
				target: "https://api.rezible.test/api/v1",
				changeOrigin: true,
				rewrite: (path) => path.replace(/^\/api/, ""),
			},
		},
	},
});
