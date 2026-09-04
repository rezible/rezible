<script lang="ts">
	import IntegrationProviderConfig from "$src/features/settings/components/integration-provider/IntegrationProviderConfig.svelte";
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";

	type Props = {
		name: string;
	};
	const { name }: Props = $props();

	registerPageDescriptor(() => ({
		title: name,
		parents: [
			{ label: "Settings", path: resolve("/settings") },
			{ label: "Integrations", path: resolve("/settings/integrations") },
		],
	}));
</script>

<IntegrationProviderConfig {name} />
