<script lang="ts">
	import type { InvestigationAttributes } from "$lib/api";
	import * as Empty from "$components/ui/empty";
	import { timestamp } from "../model";
	import { useSituationInvestigationController } from "./controller.svelte";

	type Props = {
		investigationAttributes: InvestigationAttributes;
	};

	const { investigationAttributes: attrs }: Props = $props();
	const controller = useSituationInvestigationController();
	const reportCreatedAt = $derived(timestamp(controller.report?.createdAt));
</script>

<header class="flex flex-col gap-2">
	<h2 class="text-lg font-semibold wrap-anywhere">{attrs.query || "Investigation"}</h2>
	{#if controller.report}
		<time class="text-xs text-muted-foreground tabular-nums" datetime={reportCreatedAt.iso}>
			Updated {reportCreatedAt.label}
		</time>
	{/if}
</header>

{#if controller.report}
	<p class="max-w-[75ch] whitespace-pre-wrap text-[15px] leading-6 wrap-anywhere">
		{controller.report.text || "Report text unavailable."}
	</p>
{:else}
	<Empty.Root>
		<Empty.Header>
			<Empty.Title>No report information</Empty.Title>
		</Empty.Header>
	</Empty.Root>
{/if}
