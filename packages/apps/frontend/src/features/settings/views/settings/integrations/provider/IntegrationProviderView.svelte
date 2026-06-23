<script lang="ts">
	import IntegrationProviderConfig from "$src/features/settings/components/integration-provider/IntegrationProviderConfig.svelte";
	import { setPageBreadcrumbs } from "$lib/app-shell.svelte";
	import { page } from "$app/state";

	type Props = {
		name: string;
	};
	const { name }: Props = $props();

	setPageBreadcrumbs(() => ([
		{ label: "Settings", path: "/settings" },
		{ label: "Integrations", path: "/settings/integrations"},
		{ label: name, path: "/settings/integrations"},
	]));
</script>

<IntegrationProviderConfig {name} />