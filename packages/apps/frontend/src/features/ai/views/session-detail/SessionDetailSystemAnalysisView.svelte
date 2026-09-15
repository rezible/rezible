<script lang="ts">
	import { initSystemAnalysisController, SystemAnalysis } from "$components/system-analysis";

	type Props = { analysisId: string };
	let { analysisId }: Props = $props();

	initSystemAnalysisController(() => analysisId);
</script>

<SystemAnalysis />
