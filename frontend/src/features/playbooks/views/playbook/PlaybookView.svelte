<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import { PlaybookViewState, playbookViewStateCtx } from "./viewState.svelte";

	type Props = {
		id: string;
	};
	const { id }: Props = $props();

	const viewState = new PlaybookViewState(() => id);
	playbookViewStateCtx.set(viewState);

	appShell.setPageBreadcrumbs(() => [
		{ label: "Playbooks" },
		{ label: viewState.playbookTitle, href: `/playbooks/${id}` },
	]);
</script>

<div class="flex flex-1 items-center justify-around">
	<div class="w-42 h-full border">
		hello
	</div>
</div>