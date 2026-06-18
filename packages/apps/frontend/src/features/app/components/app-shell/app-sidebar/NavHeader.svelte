<script lang="ts">
	import { useUserSessionState } from "$src/lib/user-session.svelte";
	import { cn } from "$lib/utils";
    import * as Sidebar from "$components/ui/sidebar";

	const session = useUserSessionState();
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
                    <img src="/images/logo.svg" alt="logo" class={cn("fill-neutral size-10")} />
                    <span data-open={sidebar.open ? true : undefined} class="hidden data-open:inline">Rezible</span>
                </a>
            {/snippet}
        </Sidebar.MenuButton>
    </Sidebar.MenuItem>
</Sidebar.Menu>
