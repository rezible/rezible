import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, loadEnv } from "vite";

const envPrefix = "";

export default defineConfig(({mode}) => {
    const env = loadEnv(mode, process.cwd(), envPrefix);
    const host = "0.0.0.0";
    const port = Number(env.REZ_APP_PORT);
    return {
        plugins: [sveltekit(), tailwindcss()],
        server: {
            host,
            port,
            strictPort: true,
            allowedHosts: [env.APP_HOST],
            proxy: {
                "/api": {
                    target: `http://${env.PROXY_BACKEND_UPSTREAM_HOST}`,
                    rewrite: (path) => path.replace(/^\/api/, ""),
                },
                "/auth": {
                    target: `http://${env.PROXY_AUTH_UPSTREAM_HOST}`,
                    changeOrigin: false,
                    secure: false,
                    // rewrite: (path) => path.replace(/^\/auth/, ""),
                }
            },
        },
    }
});
