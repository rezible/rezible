import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, loadEnv } from "vite";

const envPrefix = "";

export default defineConfig(({mode}) => {
    const env = loadEnv(mode, process.cwd(), envPrefix);
    const host = "0.0.0.0";
    const port = Number(env.REZ_APP_PORT);
    console.log(`listening on ${host}:${port}`);
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
                "/dex": {
                    target: `http://${env.PROXY_DEX_UPSTREAM_HOST}`,
                    // rewrite: (path) => path.replace(/^\/dex/, ""),
                }
            },
        },
    }
});
