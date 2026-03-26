import path from "path";
import adapterStatic from "@sveltejs/adapter-static";
import adapterNode from "@sveltejs/adapter-node";

const getAdapter = () => {
	const adapter = process.env.SVELTEKIT_ADAPTER;
	if (adapter === "node") {
		return adapterNode({
			out: 'build',
            precompress: true
		});
	}
	return adapterStatic({
		pages: "dist",
		assets: "dist",
		fallback: "index.html",
		precompress: false,
		strict: false,
	})
}

/** @type {import('@sveltejs/kit').Config} */
const config = {
	compilerOptions: {},
	kit: {
		output: {
			bundleStrategy: "inline",
		},
		adapter: getAdapter(),
		alias: {
			$src: path.resolve("./src"),
			$lib: path.resolve("./src/lib"),
			$components: path.resolve("./src/components"),
			$features: path.resolve("./src/features"),
			$params: path.resolve("./src/params"),
		},
	},
};

export default config;
