import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, type PluginOption } from "vite";

export default defineConfig(({ command }) => {
	const plugins: PluginOption[] = [tailwindcss(), sveltekit()];
	if (command === "build") return { plugins };

	plugins.push({
		name: "validate-environment",
		configureServer() {
			if (!process.env.APP_URL) throw new Error("APP_URL is required");

			for (const name of ["APP_PORT", "BACKEND_PORT"]) {
				const port = Number(process.env[name]);
				if (!Number.isInteger(port) || port < 1 || port > 65535) {
					throw new Error(`${name} must be a valid TCP port`);
				}
			}
		},
	});

	const appHostname = process.env.APP_URL ? new URL(process.env.APP_URL).hostname : undefined;
	return {
		plugins,
		server: {
			host: "127.0.0.1",
			port: Number(process.env.APP_PORT),
			strictPort: true,
			allowedHosts: appHostname ? [appHostname] : [],
			proxy: {
				"/api": {
					target: `http://localhost:${process.env.BACKEND_PORT}`,
					rewrite: (path) => path.replace(/^\/api/, ""),
				},
			},
		},
	};
});
