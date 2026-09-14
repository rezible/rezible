<script lang="ts">
	import { resolve } from "$app/paths";
	import * as Sheet from "$components/ui/sheet";
	import { Button } from "$components/ui/button";
	import { useSituationBriefController } from "./controller.svelte";

	const controller = useSituationBriefController();
	const record = $derived(controller.inspectedRecord);
</script>

<Sheet.Root bind:open={controller.sourceSheetOpen}>
	<Sheet.Content class="overflow-y-auto sm:max-w-lg" onCloseAutoFocus={controller.restoreSourceFocus}>
		<Sheet.Header>
			<Sheet.Title>{record?.title || "Source record"}</Sheet.Title>
			<Sheet.Description>Source in observation group: {controller.inspectedGroup}</Sheet.Description>
		</Sheet.Header>
		{#if record}
			<dl class="flex flex-col gap-4 px-4 pb-4 text-sm">
				<div>
					<dt class="text-muted-foreground">Identity</dt>
					<dd class="font-mono text-xs wrap-anywhere">{record.id}</dd>
				</div>
				<div>
					<dt class="text-muted-foreground">Type</dt>
					<dd>{record.type}</dd>
				</div>
				<div>
					<dt class="text-muted-foreground">{record.timeLabel}</dt>
					<dd><time datetime={record.time.iso}>{record.time.label}</time></dd>
				</div>
				<div>
					<dt class="text-muted-foreground">Source content</dt>
					<dd class="whitespace-pre-wrap wrap-anywhere">
						{record.content || "Source content unavailable."}
					</dd>
				</div>
				{#each record.fields as field (field.label)}
					<div>
						<dt class="text-muted-foreground">{field.label}</dt>
						<dd class="whitespace-pre-wrap wrap-anywhere">{field.value}</dd>
					</div>
				{/each}
			</dl>
			<Sheet.Footer>
				{#each record.links as link (link.label)}
					<Button variant="outline" href={link.href} target="_blank" rel="noopener noreferrer"
						>{link.label}</Button
					>
				{:else}
					<p class="text-xs text-muted-foreground">Original-source URL unavailable.</p>
				{/each}
				{#if record.eventId}
					<Button variant="outline" href={resolve("/events/[id]", { id: record.eventId })}
						>Open event</Button
					>
				{/if}
				{#if record.definitionId}
					<Button
						variant="outline"
						href={resolve("/signals/[id]/[[view=signalView]]", { id: record.definitionId })}
						>Open signal</Button
					>
				{/if}
			</Sheet.Footer>
		{/if}
	</Sheet.Content>
</Sheet.Root>
