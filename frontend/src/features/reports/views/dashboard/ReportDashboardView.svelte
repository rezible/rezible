<script lang="ts">
	import { Header } from "svelte-ux";
	import IncidentsGraph from "$features/reports/components/incidents-graph/IncidentsGraph.svelte";
</script>

<div class="">
	<Header title="Incidents by Weekday" />
	<IncidentsGraph />
</div>
