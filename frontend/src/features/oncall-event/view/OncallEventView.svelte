<script lang="ts">
	import { appShell } from "$features/app-shell/lib/appShellState.svelte";
	import { EventViewState, eventViewStateCtx } from "./viewState.svelte";

	type Props = {
		id: string;
	};
	const { id }: Props = $props();

	const viewState = new EventViewState(() => id);
	eventViewStateCtx.set(viewState);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Events" },
		{ label: viewState.eventTitle, href: `/events/${id}` },
	]);
</script>

<div class="flex gap-2 p-2">
	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Details</span>

		<!-- <span>Description</span> -->

		<!-- <span>Owning Team</span> -->
	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Roster</span>

	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Alert</span>

	</div>

	<div class="flex flex-col gap-2 w-fit border p-2">
		<span class="uppercase font-semibold text-surface-content/90">Annotations</span>

	</div>
</div>