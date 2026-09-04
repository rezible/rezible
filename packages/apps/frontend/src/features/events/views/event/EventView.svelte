<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initEventViewController } from "./controller.svelte";

	const { id }: IdProp = $props();

	const view = initEventViewController(() => id);

	registerPageDescriptor(() => ({
		title: view.eventTitle || "Event",
		parents: [{ label: "Events", path: resolve("/events") }],
	}));
</script>

<div class="flex gap-2 p-2">
	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Details</span>
	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Alert</span>
	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Annotations</span>
	</div>
</div>
