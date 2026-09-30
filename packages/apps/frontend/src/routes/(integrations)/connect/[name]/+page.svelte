<script lang="ts">
	import type { PageProps } from "./$types";
	import ConnectIntegrationView from "$features/settings/views/connect/ConnectIntegrationView.svelte";

	const { params }: PageProps = $props();
</script>

<ConnectIntegrationView name={params.name} />
