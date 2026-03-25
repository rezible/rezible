<script lang="ts">
	import { useAuthSessionState } from "$lib/auth.svelte";
	import UserMenu from "./UserMenu.svelte";

    type Props = {
        showText?: boolean;
    };
    const { showText = true }: Props = $props();

	const session = useAuthSessionState();
    const preloadHome = $derived(session.error ? "tap" : "hover");
</script>

<nav class="w-dvw h-16 max-h-16 flex border flex-1 items-center justify-between px-4">
    <a href="/" class="text-2xl flex items-center" 
        data-sveltekit-preload-data={preloadHome} 
        data-sveltekit-preload-code={preloadHome}
    >
        <img src="/images/logo.svg" alt="logo" class="h-10 w-10 fill-neutral" />
        <span>Rezible</span>
    </a>

    {#if session.isAuthenticated}
        <div class="grid place-items-center">
            <UserMenu />
        </div>
    {/if}
</nav>