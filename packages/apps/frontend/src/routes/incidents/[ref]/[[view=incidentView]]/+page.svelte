<script lang="ts">
	import type { PageProps } from "./$types";
	import IncidentView from "$features/incidents/views/incident";

	const { params }: PageProps = $props();
</script>

{#key params.ref}
	<IncidentView ref={params.ref} />
{/key}
