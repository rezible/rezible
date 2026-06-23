<script lang="ts">
	import ConnectIntegrationView from "$features/settings/views/connect/ConnectIntegrationView.svelte";
	import type { PageProps } from "./$types";

	const { params }: PageProps = $props();
</script>

<ConnectIntegrationView name={params.name} />
