<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";

	registerPageDescriptor(() => ({
		title: "Scheduled",
		parents: [{ label: "Meetings", path: resolve("/meetings") }],
	}));
</script>

<span>scheduled meetings</span>
