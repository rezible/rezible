<script lang="ts">
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import TiptapEditor, { Editor as SvelteEditor } from "$components/tiptap-editor/TiptapEditor.svelte";
	import { PlaybookViewState, playbookViewStateCtx } from "./viewState.svelte";
	import { Button } from "svelte-ux";
	import PlaybookEditor from "./PlaybookEditor.svelte";
	import PlaybookPageActions from "./PlaybookPageActions.svelte";

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
	appShell.setPageActions(PlaybookPageActions, false, () => ({viewState}));
</script>

<div class="flex flex-col h-full w-2/3 items-center self-center">
	{#if viewState.editing}
		<PlaybookEditor />
	{:else}
		<div class="flex-1 min-h-0 w-full">
			{@html viewState.playbookContent}
		</div>
	{/if}
</div>