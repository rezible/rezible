<script lang="ts">
	import type { PageProps } from "./$types";
	import SituationView from "$features/situations/views/situation";

	const { params }: PageProps = $props();
</script>

<SituationView id={params.id} />
