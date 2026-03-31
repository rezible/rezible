import eslint from '@eslint/js';
import tseslint from '@typescript-eslint/eslint-plugin';
import tsParser from '@typescript-eslint/parser';
import sveltePlugin from 'eslint-plugin-svelte';
import svelteParser from 'svelte-eslint-parser';
import unusedImports from "eslint-plugin-unused-imports";
import globals from 'globals';

const jsGlobals = {
	...globals.browser,
	...globals.es2017,
	...globals.node
}

const ignores = [
	'node_modules/**',
	'dist/**',
	'build/**',
	'.svelte-kit/**',
	'package/**',
	'.env',
	'.env.*',
	'!.env.example',
	"**/*.gen**"
];

const tsRules = {
	...tseslint.configs.recommended.rules,
	'@typescript-eslint/no-explicit-any': 'warn',
	'no-unused-vars': "off",
	"@typescript-eslint/no-unused-vars": "off",
	"unused-imports/no-unused-imports": "error",
	"unused-imports/no-unused-vars": [
		"warn",
		{
			"vars": "all",
			"varsIgnorePattern": "^_",
			"args": "after-used",
			"argsIgnorePattern": "^_",
		},
	]
}

export default [
	{ignores},
	{
		// JavaScript files
		files: ['**/*.js'],
		...eslint.configs.recommended,
		languageOptions: {
			ecmaVersion: 'latest',
			sourceType: 'module',
			globals: jsGlobals,
		}
	},
	{
		// TypeScript files
		files: ['**/*.ts'],
		plugins: {
			'@typescript-eslint': tseslint,
			"unused-imports": unusedImports,
		},
		languageOptions: {
			parser: tsParser,
			parserOptions: {
				ecmaVersion: 'latest',
				sourceType: 'module',
				project: './tsconfig.json'
			},
			globals: jsGlobals,
		},
		rules: tsRules,
	},
	{
		// Svelte files
		files: ['**/*.svelte'],
		plugins: {
			svelte: sveltePlugin,
			"@typescript-eslint": tseslint,
			"unused-imports": unusedImports
		},
		languageOptions: {
			parser: svelteParser,
			parserOptions: {
				parser: {
					ts: tsParser,
					js: null
				},
				extraFileExtensions: ['.svelte'],
				project: ['./tsconfig.json']
			}
		},
		rules: {
			...sveltePlugin.configs.recommended.rules,
			...tsRules,
		}
	}
];