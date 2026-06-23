<script lang="ts">
	import SettingsLayout from "$features/settings/views/settings/SettingsLayout.svelte";
	import type { LayoutProps } from "./$types";

	const { children }: LayoutProps = $props();
</script>

<SettingsLayout>
	{@render children()}
</SettingsLayout>
