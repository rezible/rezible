<script lang="ts">
	import type { InvestigationAttributes } from "$lib/api";
	import { SystemAnalysis, initSystemAnalysisController } from "$components/system-analysis";
	import { timestamp } from "../model";
	import { Button } from "$components/ui/button";
	import { investigationHref } from "$features/situations/lib/routes";
	import { page } from "$app/state";

	type Props = {
		situationId: string;
		investigationAttributes: InvestigationAttributes;
	};
	const { situationId, investigationAttributes }: Props = $props();

	const analysis = initSystemAnalysisController(
		() => investigationAttributes.analysisId,
		() => ({ readOnly: true })
	);

	const updatedAt = $derived(timestamp(investigationAttributes.updatedAt));

	const reportHref = $derived(investigationHref(situationId, page.url.search));
</script>

<header class="flex flex-wrap items-start justify-between gap-3">
	<div class="flex min-w-0 flex-col gap-1">
		<h2 class="text-lg font-semibold wrap-anywhere">
			{investigationAttributes.query || "Investigation"}
		</h2>
		<time class="text-xs text-muted-foreground" datetime={updatedAt.iso}>
			Updated {updatedAt.label}
		</time>
	</div>
	<div class="flex gap-2">
		<Button variant="outline" size="sm" onclick={analysis.refreshAll}>Refresh analysis</Button>
		<Button variant="outline" size="sm" href={reportHref}>View report</Button>
	</div>
</header>
<div class="min-h-0 flex-1">
	<SystemAnalysis />
</div>
