<script lang="ts">
	import ReportView from "$features/reports/views/report-view/ReportView.svelte";
	import ReportBuilder from "$features/reports/views/report-builder/ReportBuilder.svelte";

	type Props = {
		id: string;
		builder: boolean;
	}
	const {id, builder}: Props = $props();
</script>

{#if id && !builder}
	<ReportView {id} />
{:else}
	<ReportBuilder />
{/if}