<script lang="ts">
	import ReportView from "$features/report/view/ReportView.svelte";
	import ReportBuilder from "$features/report-builder/view/ReportBuilder.svelte";
	import type { PageData } from "./$types";

	type Props = {
		data: PageData;
	}
	const {data}: Props = $props();
</script>

{#if data.id && !data.builder}
	<ReportView id={data.id} />
{:else}
	<ReportBuilder />
{/if}