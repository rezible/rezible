<script lang="ts">
	import type { PageProps } from "./$types";
	import IncidentView from "$features/incidents/views/incident";

	const { data }: PageProps = $props();
</script>

<IncidentView slug={data.slug} />
