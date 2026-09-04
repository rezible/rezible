<script lang="ts">
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initPlaybookViewController } from "./controller.svelte";
	import PlaybookEditor from "./PlaybookEditor.svelte";
	import PlaybookPageActions from "./PlaybookPageActions.svelte";

	const { id }: IdProp = $props();

	const view = initPlaybookViewController(() => id);

	registerPageDescriptor(() => ({
		title: view.playbookTitle || "Playbook",
		parents: [{ label: "Playbooks", path: resolve("/playbooks") }],
		actions: { component: PlaybookPageActions, props: { view } },
	}));
</script>

<div class="flex gap-4 h-full w-full justify-between">
	<div class="flex flex-col w-3/5 max-w-4xl items-center">
		{#if view.editing}
			<PlaybookEditor />
		{:else}
			<div class="flex-1 min-h-0 w-full p-2 border">
				{@html view.playbookContent}
			</div>
		{/if}
	</div>

	<!-- <div class="flex flex-col h-full self-end border p-2 w-2/5 max-w-md">
		<Header title="Playbook Details"></Header>

		<PlaybookAlerts />
	</div> -->
</div>
