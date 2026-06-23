<script lang="ts">
	import ProviderIntegrations from "$src/features/settings/views/settings/integrations/provider/IntegrationProviderView.svelte";
	import type { PageProps } from "./$types";

	const { params }: PageProps = $props();
</script>

<ProviderIntegrations name={params.provider} />
