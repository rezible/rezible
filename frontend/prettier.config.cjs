/**
 * @see https://prettier.io/docs/en/configuration.html
 * @type {import("prettier").Config}
 */
module.exports = {
	plugins: ["prettier-plugin-svelte"],
	overrides: [{ files: "*.svelte", options: { parser: "svelte" } }],
	trailingComma: "es5",
	tabWidth: 4,
	semi: true,
	singleQuote: false,
	useTabs: true,
	printWidth: 110,
};
