<script lang="ts">
	import { Button } from "$components/ui/button";
	import * as Alert from "$components/ui/alert";
	import * as Empty from "$components/ui/empty";
	import * as Collapsible from "$components/ui/collapsible";
	import RiArrowDownSLine from "remixicon-svelte/icons/arrow-down-s-line";
	import RiFileListLine from "remixicon-svelte/icons/file-list-line";
	import RiInformationLine from "remixicon-svelte/icons/information-line";
	import { useSituationController } from "../controller.svelte";
	import { initSituationBriefController } from "./controller.svelte";
	import SituationSourceSheet from "./SituationSourceSheet.svelte";
	import { timestamp } from "../model";

	const controller = useSituationController();
	const attrs = $derived(controller.situation?.attributes);

	const brief = initSituationBriefController();
	const previewAttrs = $derived(brief.preview?.attributes);
	let fallbackTarget = $state<HTMLElement>();
	$effect(() => brief.setFallbackTarget(fallbackTarget));
</script>

{#if attrs}
	<div class="min-h-0 min-w-0 flex-1 overflow-y-auto p-4" bind:this={fallbackTarget} tabindex="-1">
		<div class="mx-auto flex max-w-6xl flex-col gap-6">
			<section class="flex max-w-[75ch] flex-col gap-3">
				<h1 class="text-[28px] leading-9 font-semibold wrap-anywhere">{attrs?.title}</h1>
				<p class="whitespace-pre-wrap text-sm leading-relaxed text-muted-foreground wrap-anywhere">
					{attrs?.summary || "Summary unavailable."}
				</p>
			</section>
			<section
				aria-labelledby="report-preview-title"
				class="flex flex-col gap-4 rounded-lg border bg-card p-4 md:p-6"
			>
				<div class="flex flex-wrap items-center justify-between gap-2">
					<h2 id="report-preview-title" class="text-lg font-semibold">Latest available report</h2>
					{#if previewAttrs}
						<time
							class="text-xs text-muted-foreground tabular-nums"
							datetime={timestamp(previewAttrs.updatedAt).iso}
							>Updated {timestamp(previewAttrs.updatedAt).label}</time
						>
					{/if}
				</div>
				{#if previewAttrs}
					<p
						class="max-w-[75ch] line-clamp-4 whitespace-pre-wrap text-[15px] leading-6 wrap-anywhere"
					>
						{previewAttrs.report?.text || "Report text unavailable."}
					</p>
					{#if brief.previewChanged}
						<Alert.Root role="note">
							<RiInformationLine />
							<Alert.Title>
								<span class="text-status-warning-foreground">
									Evidence has changed since this report.
								</span>
							</Alert.Title>
						</Alert.Root>
					{/if}
					<Button variant="link" href={brief.previewHref} class="self-start">Read report</Button>
				{:else if attrs?.investigations.length}
					<p class="text-sm text-muted-foreground">No report available yet.</p>
					<Button variant="outline" href={brief.previewHref} class="self-start">
						Open Investigations
					</Button>
				{:else}
					<p class="text-sm text-muted-foreground">No investigation yet.</p>
					<Button href={brief.previewHref} class="self-start">Run investigation</Button>
				{/if}
			</section>
			
			<section aria-labelledby="observations-title" class="flex flex-col gap-3">
				<div class="flex flex-wrap items-center justify-between gap-2">
					<div class="flex flex-wrap items-baseline gap-3">
						<h2 id="observations-title" class="text-lg font-semibold">Observations</h2>
						<p class="text-xs text-muted-foreground tabular-nums">
							{brief.observations.items.length} groups · {brief.observations.sourceCount} sources
						</p>
					</div>
					{#if brief.observations.items.length}
						<div class="flex gap-1">
							<Button variant="ghost" size="sm" onclick={() => brief.setAllGroups(true)}>
								Expand all
							</Button>
							<Button variant="ghost" size="sm" onclick={() => brief.setAllGroups(false)}>
								Collapse all
							</Button>
						</div>
					{/if}
				</div>
				{#each brief.observations.items as group (group.id)}
					<Collapsible.Root
						open={brief.groupOpen(group.id)}
						onOpenChange={(open) => brief.setGroupOpen(group.id, open)}
						class="rounded-lg border bg-card"
					>
						<Collapsible.Trigger
							class="group flex w-full items-center gap-3 p-4 text-left text-sm font-medium hover:bg-accent focus-visible:outline-2 focus-visible:outline-ring"
						>
							{#snippet child({ props })}
								<div {...props}>
									<RiFileListLine class="size-6" aria-hidden="true" />
									<span class="min-w-0 flex-1 wrap-anywhere">{group.title}</span>
									<span class="shrink-0 text-xs text-muted-foreground tabular-nums">{group.records.length} sources</span>
									<RiArrowDownSLine class="size-6 group-data-[state=open]:rotate-180" aria-hidden="true" />
								</div>
							{/snippet}
						</Collapsible.Trigger>
						<Collapsible.Content>
							{#if group.body}
								<div class="flex flex-col gap-1 border-t px-4 py-3">
									<h3 class="text-xs font-medium text-muted-foreground">
										Group interpretation
									</h3>
									<p class="max-w-[75ch] whitespace-pre-wrap text-sm wrap-anywhere">
										{group.body}
									</p>
								</div>
							{/if}
							<ul>
								{#each group.records as record (record.key)}
									<li class="flex flex-col gap-3 border-t p-4 sm:flex-row sm:items-center">
										<div class="flex min-w-0 flex-1 flex-col gap-1">
											<p class="text-[15px] font-medium wrap-anywhere">
												{record.title}
											</p>
											<p class="line-clamp-2 whitespace-pre-wrap text-sm text-muted-foreground wrap-anywhere">
												{record.content || "Source content unavailable."}
											</p>
										</div>
										<div class="flex min-w-0 flex-col gap-1 text-xs text-muted-foreground sm:w-52">
											<p class="wrap-anywhere">{record.type} · {record.source}</p>
											<time datetime={record.time.iso} class="tabular-nums">
												{record.timeLabel} {record.time.label}
											</time>
										</div>
										<Button
											variant="ghost"
											size="sm"
											class="self-start sm:self-center"
											aria-label={`Inspect ${record.title}`}
											onclick={(event) =>
												brief.inspectSource(record, group.title, event.currentTarget)}
										>Inspect</Button>
									</li>
								{:else}
									<li>
										<Empty.Root>
											<Empty.Header>
												<Empty.Title>No source records</Empty.Title>
											</Empty.Header>
										</Empty.Root>
									</li>
								{/each}
							</ul>
						</Collapsible.Content>
					</Collapsible.Root>
				{:else}
					<Empty.Root>
						<Empty.Header>
							<Empty.Title>No observations recorded</Empty.Title>
						</Empty.Header>
					</Empty.Root>
				{/each}
			</section>
		</div>
	</div>
{/if}

<SituationSourceSheet />
