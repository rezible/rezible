<script lang="ts">
	import type { PageProps } from "./$types";
	import AlertView from "$features/signals/views/signal";

	const { params }: PageProps = $props();
</script>

<AlertView id={params.id} />
