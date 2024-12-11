const svelteUx = require('svelte-ux/plugins/tailwind.cjs');

/** @type {import('tailwindcss').Config} */
module.exports = {
	content: [
		'./src/**/*.{html,js,svelte,ts}',
		'../node_modules/svelte-ux/**/*.{html,svelte,js,ts}',
		'../node_modules/layerchart/**/*.{html,svelte,js,ts}'
	],
	plugins: [
		require('@tailwindcss/typography'),
		svelteUx({colorSpace: "oklch"})
	],
	ux: {
		themes: require("./src/themes.json"),
	},
    // darkMode: 'selector',
    theme: {
        extend: {
      		gridTemplateRows: {
				'layout': 'auto 1fr',
			},
		}
    },
};
