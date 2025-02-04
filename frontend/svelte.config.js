import adapter from "@sveltejs/adapter-static";
import path from "path";

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {
		// customElement: true,
	},
	kit: {
		output: {
			bundleStrategy: "inline",
		},
		adapter: adapter({
			pages: "dist",
			assets: "dist",
			fallback: "index.html",
			precompress: false,
			strict: false,
		}),
		alias: {
			$src: path.resolve("./src"),
			$lib: path.resolve("./src/lib"),
			$components: path.resolve("./src/components"),
			$features: path.resolve("./src/features"),
		},
	},
};

export default config;
