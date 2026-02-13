import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig } from "vite";

export default defineConfig({
	plugins: [sveltekit(), tailwindcss()],
	server: {
		host: "127.0.0.1",
		port: 5173,
		strictPort: true,
		allowedHosts: ["app.dev.rezible.com"],
		proxy: {
			"/api": {
				target: "http://localhost:8888",
				changeOrigin: true,
				configure(proxy, options) {
					// console.log(options);
				},
				// rewrite: (path) => path.replace(/^\/api/, ""),
			},
		},
	},
});
