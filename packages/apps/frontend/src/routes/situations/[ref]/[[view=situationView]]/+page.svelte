<script lang="ts">
	import type { PageProps } from "./$types";
	import SituationView from "$features/situations/views/situation";
	type Props = PageProps;
	const { params }: Props = $props();
</script>

{#key params.ref}
	<SituationView ref={params.ref} />
{/key}
