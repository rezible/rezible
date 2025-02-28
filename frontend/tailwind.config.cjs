const layerstack = require("@layerstack/tailwind/plugin");

/** @type {import('tailwindcss').Config} */
module.exports = {
	content: [
		"./src/**/*.{html,js,svelte,ts}",
		"../node_modules/svelte-ux/**/*.{svelte,js}",
		"../node_modules/layerchart/**/*.{svelte,js}",
	],
	plugins: [
		layerstack({ colorSpace: "oklch" }),
	],
	ux: {
		themes: require("./src/themes.json"),
	},
	theme: {
		extend: {
			gridTemplateRows: {
				"app-shell-layout": "auto 1fr",
			},
		},
	},
};
