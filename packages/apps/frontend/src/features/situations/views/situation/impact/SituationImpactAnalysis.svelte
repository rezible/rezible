<script lang="ts">
	import type { SituationInvestigation } from "@rezible/api-client-ts";
	import { SystemAnalysis, initSystemAnalysisController } from "$components/system-analysis";
	import { timestamp } from "../model";
	import { Button } from "$components/ui/button";
	import { investigationHref } from "$features/situations/lib/routes";
	import { page } from "$app/state";

    type Props = {
        investigation: SituationInvestigation;
    }
    const { investigation }: Props = $props();

    const attrs = $derived(investigation.attributes);
	const analysis = initSystemAnalysisController(() => attrs.analysisId, () => ({ readOnly: true }));

    const updatedAt = $derived(timestamp(attrs.updatedAt));

    const reportHref = $derived(investigationHref(attrs.situationId, investigation.id, page.url.search));
</script>

<header class="flex flex-wrap items-start justify-between gap-3">
    <div class="flex min-w-0 flex-col gap-1">
        <h2 class="text-lg font-semibold wrap-anywhere">
            {attrs.query || "Investigation"}
        </h2>
        <time class="text-xs text-muted-foreground" datetime={updatedAt.iso}>
            Updated {updatedAt.label}
        </time>
    </div>
    <div class="flex gap-2">
        <Button variant="outline" size="sm" onclick={analysis.refreshAll}>
            Refresh analysis
        </Button>
        <Button variant="outline" size="sm" href={reportHref}>
            View report
        </Button>
    </div>
</header>
<div class="min-h-0 flex-1">
    <SystemAnalysis />
</div>