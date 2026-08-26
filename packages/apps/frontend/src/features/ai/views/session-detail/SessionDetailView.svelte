<script lang="ts">
	import * as Tabs from "$components/ui/tabs";
	import { Button } from "$components/ui/button";
	import { initSessionDetailController } from "./controller.svelte";
	import MessageParts from "./MessageParts.svelte";
	import type { LiveOverlay, TranscriptGroup } from "./model";
	import SessionDetailSystemAnalysisView from "./SessionDetailSystemAnalysisView.svelte";

	type Props = { sessionId: string };
	let { sessionId }: Props = $props();

	const formatTime = (value?: string) => (value ? new Date(value).toLocaleString() : "—");

	const getLiveOverlayModelsForTurn = ({models}: LiveOverlay, turnId?: string) => {
		if (!turnId) return [];
		return [...models.values()].filter((m) => m.turnId === turnId)
	};

	const controller = initSessionDetailController(() => sessionId);
	const analysisId = $derived(controller.session?.attributes.systemAnalysisId);
	const transcript = $derived(controller.transcript);
	const artifacts = $derived(controller.artifacts);
</script>

{#snippet transcriptGroup(group: TranscriptGroup)}
	{@const liveModels = getLiveOverlayModelsForTurn(controller.overlay, group.turn?.id)}
	<section class="border border-border bg-card p-3">
		{#if group.turn}
			{@const attrs = group.turn.attributes}
			{@const timeRange = `${formatTime(attrs.startedAt)} → ${formatTime(attrs.finishedAt)}`}
			{@const finishReason = attrs.finishReason ? `· ${attrs.finishReason}` : ""}
			<header class="mb-3 flex flex-wrap justify-between gap-2 text-sm">
				<strong>Turn {attrs.sequence}</strong>
				<span>
					{attrs.status} · {timeRange} {finishReason}
				</span>
			</header>
			{#if attrs.error}
				{@const turnErrJson = JSON.stringify(attrs.error, null, 2)}
				<pre class="mb-3 whitespace-pre-wrap bg-destructive/10 p-2 text-xs text-destructive">{turnErrJson}</pre>
			{/if}
		{/if}
		{#each group.messages as message (message.id)}
			{@const {attributes: attrs} = message}
			<article class="mb-2 border-l-2 border-border pl-3 text-sm">
				<div class="mb-1 text-xs text-muted-foreground">
					{attrs.role} · message {attrs.sequence}
				</div>
				<MessageParts parts={attrs.parts} />
			</article>
		{/each}

		{#each liveModels as model (`${model.turnId}:${model.index}`)}
			<article class="mb-2 border-l-2 border-primary pl-3 text-sm">
				<div class="text-xs text-muted-foreground">Live {model.role}</div>
				<MessageParts parts={model.parts} />
			</article>
		{/each}
	</section>
{/snippet}

{#if controller.isPending}
	<div class="p-6 text-muted-foreground">Loading session history…</div>
{:else if controller.error}
	<div class="p-6 text-destructive">
		Could not load this session.
		<Button variant="outline" onclick={controller.retry}>Retry</Button>
	</div>
{:else if controller.session}
	{@const {id: sessionId, attributes: attrs} = controller.session}
	<div class="flex h-full min-h-0 flex-col gap-4 p-4">
		<header class="flex flex-wrap items-start justify-between gap-3 border-b border-border pb-3">
			<div>
				<h1 class="text-xl font-semibold">{attrs.agentName}</h1>
				<div class="font-mono text-xs text-muted-foreground">{sessionId}</div>
			</div>
			<div class="text-right text-xs text-muted-foreground">
				<div>
					{controller.latestTurn?.attributes.status ?? "No turns"} · Live: {controller.connection}
				</div>
				<div>
					Created {formatTime(attrs.createdAt)} · Updated {formatTime(attrs.updatedAt)}
				</div>
			</div>
		</header>
		
		{#if controller.historyIncomplete}
			<div class="text-xs text-muted-foreground">Loading complete history…</div>
		{/if}

		<Tabs.Root value="transcript" class="flex min-h-0 grow flex-col">
			<Tabs.List>
				<Tabs.Trigger value="transcript">Transcript</Tabs.Trigger>
				<Tabs.Trigger value="artifacts">Artifacts ({controller.artifacts.length})</Tabs.Trigger>
				{#if analysisId}
					<Tabs.Trigger value="analysis">Analysis</Tabs.Trigger>
				{/if}
			</Tabs.List>

			<Tabs.Content value="transcript" class="min-h-0 grow overflow-auto">
				<div class="space-y-4 py-3">
					{#each transcript.groups as group (group.turn?.id)}
						{@render transcriptGroup(group)}
					{/each}
					{#if transcript.ungrouped.length}
						<section class="border border-destructive/40 p-3">
							<h2 class="mb-2 font-semibold">Ungrouped messages</h2>
							{#each transcript.ungrouped as message (message.id)}
								<article class="mb-2">
									<MessageParts parts={message.attributes.parts} />
								</article>
							{/each}
						</section>
					{/if}
					{#if !transcript.groups.length && !transcript.ungrouped.length}
						<p class="text-sm text-muted-foreground">No persisted history.</p>
					{/if}
				</div>
			</Tabs.Content>

			<Tabs.Content value="artifacts" class="min-h-0 grow overflow-auto py-3">
				{#each artifacts as artifact (artifact.id)}
					{@const {name, parts} = artifact.attributes}
					<article class="mb-3 border border-border p-3">
						<h2 class="mb-2 font-medium">{name}</h2>
						<MessageParts {parts} />
					</article>
				{:else}
					<p class="text-muted-foreground">No artifacts.</p>
				{/each}
			</Tabs.Content>

			{#if analysisId}
				<Tabs.Content value="analysis" class="min-h-0 grow">
					<SessionDetailSystemAnalysisView {analysisId} />
				</Tabs.Content>
			{/if}
		</Tabs.Root>
	</div>
{/if}
