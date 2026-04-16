<script lang="ts">
	import { useAuthSessionState } from "$lib/auth.svelte";
	import { cn } from "$lib/utils";
    import * as Sidebar from "$components/ui/sidebar";

	const session = useAuthSessionState();
    const sidebar = Sidebar.useSidebar();
    const preloadHome = $derived(session.error ? "tap" : "hover");
</script>

<Sidebar.Menu>
    <Sidebar.MenuItem>
        <Sidebar.MenuButton size="lg">
            {#snippet child({ props })}
                <a href="/" {...props} 
                    class="text-2xl flex items-center gap-2"
                    data-sveltekit-preload-data={preloadHome} 
                    data-sveltekit-preload-code={preloadHome}
                >
                    <img src="/images/logo.svg" alt="logo" class={cn("fill-neutral", sidebar.open ? "size-10" : "size-8")} />
                    <span data-open={sidebar.open ? true : undefined} class="hidden data-open:inline">Rezible</span>
                </a>
            {/snippet}
        </Sidebar.MenuButton>
    </Sidebar.MenuItem>
</Sidebar.Menu>

<!-- <nav class={cn("flex-1 items-center justify-between flex", sidebar.open ? "h-16 max-h-16" : "flex-col")}>
    <a href="/" class="text-2xl flex items-center gap-2 flex-1 h-10" 
        data-sveltekit-preload-data={preloadHome} 
        data-sveltekit-preload-code={preloadHome}
    >
        <img src="/images/logo.svg" alt="logo" class={cn("fill-neutral", sidebar.open ? "size-10" : "size-6")} />
        <span class={cn(sidebar.open ? "" : "hidden")}>Rezible</span>
    </a>
</nav> -->