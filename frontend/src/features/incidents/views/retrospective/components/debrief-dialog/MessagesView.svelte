<script lang="ts">
	import { Avatar as UxAvatar } from 'svelte-ux';
	import { mdiAbacus } from '@mdi/js';
	import type { IncidentDebriefMessage } from '$lib/api';
    import { session } from '$lib/auth.svelte';
	import Avatar from '$components/avatar/Avatar.svelte';
	import PulseLoader from '$components/loader/PulseLoader.svelte';
    import { onMount } from 'svelte';

	interface Props { 
		messages: IncidentDebriefMessage[],
		waitingForResponse: boolean,
	};
	let { messages, waitingForResponse }: Props = $props();

	let container = $state<HTMLElement>();
	const scrollToLatestMessage = () => {
		if (container) container.scrollIntoView({ behavior: "instant", block: "end" });
	}

	const obs = new ResizeObserver(scrollToLatestMessage);
	onMount(() => {
		// hacky but eh
		setTimeout(() => {
			if (container) {
				obs.observe(container);
				scrollToLatestMessage();
			}
		}, 1);
		return () => obs.disconnect();
	})
</script>

<div class="flex flex-col justify-end gap-2 w-full min-h-96 scroll-smooth overflow-y-auto" bind:this={container}>
	{#each messages as msg}
		{#if msg.attributes.type === "user"}
			{@render userMessage(msg)}
		{:else if msg.attributes.type === "assistant"}
			{@render assistantMessage(msg)}
		{/if}
	{/each}
	{#if waitingForResponse}
		<div class="flex gap-4 w-2/3 items-center">
			<UxAvatar class="border" icon={mdiAbacus} />
			<PulseLoader />
		</div>
	{/if}
</div>

{#snippet userMessage(msg: IncidentDebriefMessage)}
	<div class="flex gap-2 px-2 w-2/3 self-end flex-row-reverse h-fit items-center">
		<Avatar kind="user" id={session.user?.id || ''} size={40} />
		<div class="peer rounded p-2 border self-start bg-neutral">
			<span>{msg.attributes.body}</span>
		</div>
	</div>
{/snippet}

{#snippet assistantMessage(msg: IncidentDebriefMessage)}
	<div class="flex gap-2 px-2 w-2/3 border-accent border-s ">
		<UxAvatar class="border" icon={mdiAbacus} />
		<div class="rounded p-2 border self-end bg-neutral">
			<span>{msg.attributes.body}</span>
		</div>
	</div>
{/snippet}