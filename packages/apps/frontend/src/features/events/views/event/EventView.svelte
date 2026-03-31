<script lang="ts">
	import { appShell } from "$features/app";
	import type { IdProps } from "$lib/utils.svelte";
	import { initEventViewController } from "./controller.svelte";

	const { id }: IdProps = $props();

	const view = initEventViewController(() => id);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Events" },
		{ label: view.eventTitle, href: `/events/${view.eventId}` },
	]);
</script>

<div class="flex gap-2 p-2">
	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Details</span>

		<!-- <span>Description</span> -->

		<!-- <span>Owning Team</span> -->
	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Alert</span>

	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Annotations</span>

	</div>
</div>