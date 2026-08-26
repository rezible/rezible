<script lang="ts">
	import type { AgentMessagePart } from "$lib/api";
	import { previousThursday } from "date-fns";
	import MessageParts from "./MessageParts.svelte";

	type Props = { parts: AgentMessagePart[] };
	let { parts }: Props = $props();

	const json = (value: unknown) => JSON.stringify(value, null, 2);
	const isSafeMedia = (part: AgentMessagePart) => {
		if (!part.url || !part.contentType) return false;

		try {
			const url = new URL(part.url, window.location.origin);
			return ["http:", "https:", "data:"].includes(url.protocol);
		} catch {
			return false;
		}
	};
</script>

{#snippet toolRequestPart(part: AgentMessagePart)}
	{@const isRequest = part.kind === "tool-request"}
	{@const payload = json(isRequest ? part.input : part.output)}
	<details class="border border-border p-2">
		<summary>
			{isRequest ? "Tool request" : "Tool response"}: {part.name ?? "unknown"}
		</summary>
		<pre class="mt-2 overflow-auto whitespace-pre-wrap text-xs">{payload}</pre>
		{#if part.content}
			<div class="mt-2 border-t pt-2">
				<MessageParts parts={part.content} />
			</div>
		{/if}
	</details>
{/snippet}

{#snippet mediaPart(part: AgentMessagePart)}
	{#if isSafeMedia(part)}
		{@const { url, contentType: ct } = part}
		{#if ct?.startsWith("image/")}
			<img class="max-h-72 max-w-full" src={url} alt="Session media" />
		{:else if ct?.startsWith("audio/")}
			<audio controls src={url}></audio>
		{:else if ct?.startsWith("video/")}
			<video class="max-h-72 max-w-full" controls muted src={url}></video>
		{:else}
			<a class="underline" href={url} rel="noreferrer">Open media resource</a>
		{/if}
	{:else}
		<span>Unsafe media not rendered</span>
	{/if}
{/snippet}

{#each parts as part, index (`${part.kind}:${index}`)}
	{#if part.kind === "text"}
		<p class="whitespace-pre-wrap">{part.text}</p>
	{:else if part.kind === "reasoning"}
		<details class="border-l-2 border-border pl-3">
			<summary>Model-provided reasoning</summary>
			<p class="whitespace-pre-wrap text-muted-foreground">{part.text}</p>
		</details>
	{:else if part.kind === "resource"}
		<span class="break-all text-muted-foreground">Resource: {part.uri ?? "Missing URI"}</span>
	{:else if part.kind === "tool-request" || part.kind === "tool-response"}
		{@render toolRequestPart(part)}
	{:else if part.kind === "media"}
		{@render mediaPart(part)}
	{:else}
		<pre class="overflow-auto whitespace-pre-wrap text-xs">{json(part)}</pre>
	{/if}
{/each}
