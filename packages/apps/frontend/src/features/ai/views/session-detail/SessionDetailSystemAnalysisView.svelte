<script lang="ts">
	import { initSystemAnalysisController, SystemAnalysisDiagram } from "$components/system-analysis";

	type Props = { analysisId: string };
	let { analysisId }: Props = $props();

	initSystemAnalysisController(() => analysisId);
</script>

<SystemAnalysisDiagram />
