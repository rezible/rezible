<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
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

<div class="flex-1">
	{id}
</div>