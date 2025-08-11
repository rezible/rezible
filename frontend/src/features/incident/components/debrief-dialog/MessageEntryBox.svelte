<script lang="ts">
	import {
		addIncidentDebriefUserMessageMutation,
		type IncidentDebrief,
		type IncidentDebriefMessage,
	} from "$lib/api";
	import { mdiArrowRight } from "@mdi/js";
	import { useQueryClient, createMutation } from "@tanstack/svelte-query";
	import Button from "$components/button/Button.svelte";

	type Props = {
		debrief: IncidentDebrief;
		disabled?: boolean;
		lastMessage?: IncidentDebriefMessage;
		onAdded: (message: IncidentDebriefMessage) => void;
	};
	let { debrief, disabled, lastMessage, onAdded }: Props = $props();

	const addMessageMut = createMutation(() => ({
		...addIncidentDebriefUserMessageMutation(),
		onSuccess: ({ data: msg }) => onAdded(msg),
	}));

	let value = $state("");
	const sendMessage = () => {
		const msgContent = $state.snapshot(value);
		const body = { attributes: { messageContent: msgContent } };
		addMessageMut.mutate({ path: { id: debrief.id }, body });

		value = "";
	};

	const usingTool = $derived(lastMessage?.attributes.type && false);
</script>

{#if usingTool}
	<span>using tool: {usingTool}</span>
{/if}

<div class="flex flex-row gap-1">
	<textarea
		{disabled}
		class="min-h-2 max-h-20 leading-6 w-full border p-2 resize-none bg-surface-100 focus:outline-none"
		style="field-sizing: content"
		bind:value
	></textarea>

	<Button
		color="accent"
		variant="fill"
		disabled={disabled || !value}
		icon={mdiArrowRight}
		class="flex-row-reverse"
		on:click={sendMessage}
	>
		Send
	</Button>
</div>
