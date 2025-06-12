const layerstack = require("@layerstack/tailwind/plugin");

/** @type {import('tailwindcss').Config} */
module.exports = {
	content: [
		"./src/**/*.{html,js,svelte,ts}",
		"../node_modules/svelte-ux/**/*.{svelte,js}",
	],
	plugins: [
		layerstack({ colorSpace: "oklch" }),
	],
	ux: {
		themes: require("./src/svelteux-themes.json"),
	},
};
