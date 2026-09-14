<script lang="ts">
	import type { PageProps } from "./$types";
	import SituationView from "$features/situations/views/situation";
	type Props = PageProps;
	const { params }: Props = $props();
</script>

{#key params.id}
	<SituationView id={params.id} />
{/key}
