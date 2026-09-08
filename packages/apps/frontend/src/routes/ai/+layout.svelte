<script lang="ts">
	import InternalRouteGuard from "$components/layout/internal-route-guard";
	import type { LayoutProps } from "./$types";

	const { children }: LayoutProps = $props();
</script>

<InternalRouteGuard>
	{@render children()}
</InternalRouteGuard>
