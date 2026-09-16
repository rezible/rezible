<script lang="ts">
	import { useSituationInvestigationController } from "./controller.svelte";
	import type { InvestigationAttributes } from "@rezible/api-client-ts";
	import { timestamp } from "../model";

	import * as Empty from "$components/ui/empty";
	import * as Alert from "$components/ui/alert";
	import RiInformationLine from "remixicon-svelte/icons/information-line";

	type Props = {
		investigationAttributes: InvestigationAttributes;
	};
	const { investigationAttributes: attrs }: Props = $props();

	const controller = useSituationInvestigationController();
	const reportCreatedAt = $derived(timestamp(controller.report?.createdAt));
</script>

<header class="flex flex-col gap-2">
	<h2 class="text-lg font-semibold wrap-anywhere">
		{attrs.query || "Investigation"}
	</h2>
	<time class="text-xs text-muted-foreground tabular-nums" datetime={reportCreatedAt.iso}>
		Updated {reportCreatedAt.label}
	</time>
</header>

{#if controller.reportChanged}
	<Alert.Root role="note">
		<RiInformationLine /><Alert.Title>
			<span class="text-status-warning-foreground">Evidence has changed since this report.</span>
		</Alert.Title>
	</Alert.Root>
{/if}

{#if controller.report}
	<p class="max-w-[75ch] whitespace-pre-wrap text-[15px] leading-6 wrap-anywhere">
		{controller.report.text || "Report text unavailable."}
	</p>
	{#each controller.reportSections as section (section.title)}
		<section class="flex max-w-[75ch] flex-col gap-2">
			<h3 class="font-semibold">{section.title}</h3>
			{#if section.items.length === 1}
				<p class="whitespace-pre-wrap text-[15px] leading-6 wrap-anywhere">
					{section.items[0]}
				</p>
			{:else}
				<ul class="flex list-disc flex-col gap-2 pl-5 text-[15px] leading-6">
					{#each section.items as item, index (index)}
						<li class="whitespace-pre-wrap wrap-anywhere">
							{item}
						</li>
					{/each}
				</ul>
			{/if}
		</section>
	{/each}
{:else}
	<Empty.Root>
		<Empty.Header>
			<Empty.Title>No report information</Empty.Title>
		</Empty.Header>
	</Empty.Root>
{/if}
