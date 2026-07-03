<script lang="ts">
	import { Badge } from "$components/ui/badge";
	import { Spinner } from "$components/ui/spinner";
	import type { AiAgentRunSnapshot } from "$lib/api";

	import AgentRunArtifact from "./AgentRunArtifact.svelte";
	import AgentRunMessage from "./AgentRunMessage.svelte";

	type Props = {
		snapshot: AiAgentRunSnapshot;
	};

	const { snapshot }: Props = $props();

	const attrs = $derived(snapshot.attributes);
	
	const state = $derived(attrs.state);
	const messages = $derived(state?.messages || []);
	const artifacts = $derived(state?.artifacts || []);

	const status = $derived(attrs.status);
	const isPending = $derived(status === "pending");
	const isCompleted = $derived(status === "completed");
	const isFailed = $derived(status === "failed");
</script>

<article class="space-y-4 rounded border border-border bg-card p-4 text-card-foreground">
	<header class="flex min-w-0 flex-col gap-3 sm:flex-row sm:items-start sm:justify-between">
		<div class="min-w-0 space-y-1">
			<div class="flex flex-wrap items-center gap-2">
				<Badge
					variant={isFailed ? "destructive" : isCompleted ? "secondary" : "outline"}
					class={[
						"capitalize",
						isPending && "text-blue-100",
						isCompleted && "text-green-700",
					]}>{status}</Badge
				>
				{#if isPending}
					<Spinner aria-label="Snapshot pending" />
				{/if}
			</div>
			<p class="break-all text-xs text-muted-foreground">{snapshot.id}</p>
		</div>
		<div class="text-left text-xs text-muted-foreground sm:text-right">
			{attrs.created_at}
		</div>
	</header>

	<dl class="grid gap-3 text-sm sm:grid-cols-2 lg:grid-cols-4">
		<div class="min-w-0">
			<dt class="font-medium text-muted-foreground">State</dt>
			<dd class="text-foreground">
				{#if state}
					{messages.length} messages,
					{artifacts.length} artifacts
				{:else}
					No state
				{/if}
			</dd>
		</div>
		<div class="min-w-0">
			<dt class="font-medium text-muted-foreground">Finish reason</dt>
			<dd class="break-words text-foreground">{attrs.finish_reason || "Not set"}</dd>
		</div>
	</dl>

	{#if attrs.error}
		<div class="rounded border border-destructive/30 bg-destructive/10 p-3 text-sm text-destructive">
			{attrs.error}
		</div>
	{/if}

	{#if state?.custom}
		<div class="space-y-1">
			<div class="text-xs font-medium text-muted-foreground">Custom state</div>
			<pre
				class="max-h-72 overflow-auto rounded border border-border bg-background p-3 text-xs leading-relaxed text-foreground">{JSON.stringify(
					state.custom,
					null,
					2
				)}</pre>
		</div>
	{/if}

	<section class="space-y-3">
		<h3 class="text-sm font-medium text-foreground">Messages</h3>
		<div class="space-y-3">
			{#each messages as message (message)}
				<AgentRunMessage {message} />
			{:else}
				<p class="text-sm text-muted-foreground">No messages.</p>
			{/each}
		</div>
	</section>

	<section class="space-y-3">
		<h3 class="text-sm font-medium text-foreground">Artifacts</h3>
		<div class="space-y-3">
			{#each artifacts as artifact, index (artifact)}
				<AgentRunArtifact {artifact} {index} />
			{:else}
				<p class="text-sm text-muted-foreground">No artifacts.</p>
			{/each}
		</div>
	</section>
</article>
