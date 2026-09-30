import { plugin } from "bun";

// Unit tests cover pure TypeScript modules. Svelte components they import
// (for example Remix icons in status presenters) load as inert stubs. The test
// script passes `--conditions svelte` so component packages resolve to their
// `.svelte` entry points.
plugin({
	name: "svelte-component-stub",
	setup(build) {
		build.onLoad({ filter: /\.svelte$/ }, (args) => ({
			contents: `export default function SvelteComponentStub() {}\nexport const stubPath = ${JSON.stringify(args.path)};`,
			loader: "js",
		}));
	},
});
