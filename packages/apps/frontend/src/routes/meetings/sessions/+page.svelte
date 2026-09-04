<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";

	registerPageDescriptor(() => ({
		title: "Sessions",
		parents: [{ label: "Meetings", path: resolve("/meetings") }],
	}));
</script>

<span> meeting sessions </span>
