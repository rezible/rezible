import eslint from "@eslint/js";
import tseslint from "@typescript-eslint/eslint-plugin";
import tsParser from "@typescript-eslint/parser";
import sveltePlugin from "eslint-plugin-svelte";
import svelteParser from "svelte-eslint-parser";
import globals from "globals";

const jsGlobals = {
	...globals.browser,
	...globals.es2017,
	...globals.node,
};

const ignores = [
	"node_modules/**",
	"dist/**",
	"build/**",
	".svelte-kit/**",
	".cache/**",
	"package/**",
	".env",
	".env.*",
	"!.env.example",
	"**/*.gen**",
];

// These rules do not require type information. Keep project parsing disabled;
// svelte-check owns type checking without making each lint run load the project.
const tsRules = {
	...tseslint.configs.recommended.rules,
	"@typescript-eslint/no-explicit-any": "warn",

	// disabled for now, too noisy
	"@typescript-eslint/no-unused-vars": "off", /*[
		"error",
		{
			vars: "all",
			varsIgnorePattern: "^_",
			args: "after-used",
			argsIgnorePattern: "^_",
		},
	],*/
};

export default [
	{ ignores },
	{
		// JavaScript files
		files: ["**/*.js"],
		...eslint.configs.recommended,
		languageOptions: {
			ecmaVersion: "latest",
			sourceType: "module",
			globals: jsGlobals,
		},
	},
	{
		// TypeScript files
		files: ["**/*.ts"],
		plugins: {
			"@typescript-eslint": tseslint,
		},
		languageOptions: {
			parser: tsParser,
			parserOptions: {
				ecmaVersion: "latest",
				sourceType: "module",
			},
			globals: jsGlobals,
		},
		rules: tsRules,
	},
	{
		// Svelte files
		files: ["**/*.svelte"],
		plugins: {
			svelte: sveltePlugin,
			"@typescript-eslint": tseslint,
		},
		languageOptions: {
			parser: svelteParser,
			parserOptions: {
				parser: {
					ts: tsParser,
					js: null,
				},
				extraFileExtensions: [".svelte"],
			},
		},
		rules: {
			...sveltePlugin.configs.recommended.rules,
			...tsRules,
		},
	},
];
