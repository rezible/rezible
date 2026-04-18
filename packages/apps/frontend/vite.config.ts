import { sveltekit } from "@sveltejs/kit/vite";
import tailwindcss from "@tailwindcss/vite";
import { defineConfig, loadEnv } from "vite";

const envPrefix = "";

export default defineConfig(({mode}) => {
    const env = loadEnv(mode, process.cwd(), envPrefix);
    const host = "0.0.0.0";
    const port = Number(env.PORT ?? env.APP_PORT ?? "7000");
    return {
        plugins: [tailwindcss(), sveltekit()],
        server: {
            host,
            port,
            strictPort: true,
            allowedHosts: [env.APP_DOMAIN],
            proxy: {
                "/api": {
                    target: `http://${env.PROXY_BACKEND_UPSTREAM_HOST}`,
                    rewrite: (path) => path.replace(/^\/api/, ""),
                }
            },
        },
    }
});
