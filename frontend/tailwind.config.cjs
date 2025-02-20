const colors = require('tailwindcss/colors');
const layerstack = require("@layerstack/tailwind/plugin");

/** @type {import('tailwindcss').Config} */
module.exports = {
	content: [
		"./src/**/*.{html,js,svelte,ts}",
		"../node_modules/svelte-ux/**/*.{svelte,js}",
		"../node_modules/layerchart/**/*.{svelte,js}",
	],
	plugins: [
		require("@tailwindcss/typography"),
		layerstack({ colorSpace: "oklch" }),
	],
	ux: {
		themes: require("./src/themes.json"),
	},
	// darkMode: 'selector',
	theme: {
		extend: {
			gridTemplateRows: {
				layout: "auto 1fr",
			},
		},
	},
};
