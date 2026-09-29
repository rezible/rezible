<script lang="ts">
	import { resolve } from "$app/paths";
	import * as Sheet from "$components/ui/sheet";
	import { Button } from "$components/ui/button";
	import { Spinner } from "$components/ui/spinner";
	import { initSituationSourceSheetController } from "./controller.svelte";
	import type { SourceTarget } from "$features/situations/lib/model";

	type Props = {
		target: SourceTarget | undefined;
		onClose: () => void;
	};

	let { target, onClose }: Props = $props();

	const controller = initSituationSourceSheetController(() => target);
	
	const directRecord = $derived(controller.directRecord);
	const evidence = $derived(controller.evidence);
	const evidenceAttributes = $derived(controller.evidenceAttributes);

	function closeWhenDismissed(open: boolean) {
		if (!open && target) {
			onClose();
		}
	}
</script>

<Sheet.Root open={target !== undefined} onOpenChange={closeWhenDismissed}>
	<Sheet.Content class="overflow-y-auto sm:max-w-lg">
		<Sheet.Header>
			<Sheet.Title>{controller.title}</Sheet.Title>
			<Sheet.Description>{controller.description}</Sheet.Description>
		</Sheet.Header>

		{#if directRecord}
			<dl class="flex flex-col gap-4 px-4 pb-4 text-sm">
				<div>
					<dt class="text-muted-foreground">Identity</dt>
					<dd class="font-mono text-xs wrap-anywhere">{directRecord.id}</dd>
				</div>
				<div>
					<dt class="text-muted-foreground">Type</dt>
					<dd>{directRecord.type}</dd>
				</div>
				<div>
					<dt class="text-muted-foreground">{directRecord.timeLabel}</dt>
					<dd><time datetime={directRecord.time.iso}>{directRecord.time.label}</time></dd>
				</div>
				<div>
					<dt class="text-muted-foreground">Source content</dt>
					<dd class="whitespace-pre-wrap wrap-anywhere">
						{directRecord.content || "Source content unavailable."}
					</dd>
				</div>
				{#each directRecord.fields as field (field.label)}
					<div>
						<dt class="text-muted-foreground">{field.label}</dt>
						<dd class="whitespace-pre-wrap wrap-anywhere">{field.value}</dd>
					</div>
				{/each}
			</dl>
			<Sheet.Footer>
				{#each directRecord.links as link (link.label)}
					<Button variant="outline" href={link.href} target="_blank" rel="noopener noreferrer">
						{link.label}
					</Button>
				{:else}
					<p class="text-xs text-muted-foreground">Original-source URL unavailable.</p>
				{/each}
				{#if directRecord.eventId}
					<Button variant="outline" href={resolve("/events/[id]", { id: directRecord.eventId })}>
						Open event
					</Button>
				{/if}
				{#if directRecord.definitionId}
					<Button
						variant="outline"
						href={resolve("/signals/[id]/[[view=signalView]]", { id: directRecord.definitionId })}
					>
						Open signal
					</Button>
				{/if}
			</Sheet.Footer>
		{:else if target?.kind === "knowledgeEvidence"}
			{#if controller.evidenceQuery.isPending && !evidence}
				<div role="status" class="flex items-center gap-2 px-4 py-6 text-sm text-muted-foreground">
					<Spinner /> Loading cited evidence
				</div>
			{:else if controller.evidenceUnavailable || (!evidence && controller.evidenceQuery.isError)}
				<div role="alert" class="flex flex-col gap-3 px-4 py-6 text-sm">
					<p>Evidence is unavailable.</p>
					{#if !controller.evidenceUnavailable}
						<Button
							variant="outline"
							size="sm"
							onclick={() => controller.evidenceQuery.refetch()}
						>
							Retry
						</Button>
					{/if}
				</div>
			{:else if evidenceAttributes}
				{#if controller.evidenceQuery.isError}
					<div
						role="status"
						class="flex items-center justify-between gap-3 px-4 text-xs text-muted-foreground"
					>
						<span>Refresh failed. Showing previously loaded evidence.</span>
						<Button variant="ghost" size="sm" onclick={() => controller.evidenceQuery.refetch()}>
							Retry
						</Button>
					</div>
				{/if}
				<dl class="flex flex-col gap-4 px-4 pb-4 text-sm">
					<div>
						<dt class="text-muted-foreground">Assertion</dt>
						<dd class="whitespace-pre-wrap wrap-anywhere">
							{evidenceAttributes.assertion || "Assertion unavailable."}
						</dd>
					</div>
					<div>
						<dt class="text-muted-foreground">Kind</dt>
						<dd>{controller.evidenceKindLabel}</dd>
					</div>
					<div>
						<dt class="text-muted-foreground">Effective time</dt>
						<dd>
							<time datetime={controller.evidenceTime.iso}>
								{controller.evidenceTime.label}
							</time>
						</dd>
					</div>
					{#if evidenceAttributes.subjectState.displayName}
						<div>
							<dt class="text-muted-foreground">Subject</dt>
							<dd class="wrap-anywhere">{evidenceAttributes.subjectState.displayName}</dd>
						</div>
					{/if}
					{#if evidenceAttributes.subjectState.description}
						<div>
							<dt class="text-muted-foreground">Subject description</dt>
							<dd class="whitespace-pre-wrap wrap-anywhere">
								{evidenceAttributes.subjectState.description}
							</dd>
						</div>
					{/if}
				</dl>
			{/if}
		{/if}

		<Sheet.Footer>
			<Button variant="outline" onclick={onClose}>Close</Button>
		</Sheet.Footer>
	</Sheet.Content>
</Sheet.Root>
