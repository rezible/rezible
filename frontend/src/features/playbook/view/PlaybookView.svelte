<script lang="ts">
	import { appShell } from "$features/app-shell";
	import { usePlaybookViewState } from "$features/playbook";
	import PlaybookEditor from "./PlaybookEditor.svelte";
	import PlaybookPageActions from "./PlaybookPageActions.svelte";

	const view = usePlaybookViewState();

	appShell.setPageBreadcrumbs(() => [
		{ label: "Playbooks" },
		{ label: view.playbookTitle, href: `/playbooks/${view.playbookId}` },
	]);
	appShell.setPageActions(PlaybookPageActions, false, () => ({viewState: view}));
</script>

<div class="flex flex-col h-full w-2/3 items-center self-center">
	{#if view.editing}
		<PlaybookEditor />
	{:else}
		<div class="flex-1 min-h-0 w-full">
			{@html view.playbookContent}
		</div>
	{/if}
</div>