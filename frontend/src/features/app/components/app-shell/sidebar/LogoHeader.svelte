<script lang="ts">
	import { useAuthSessionState } from "$lib/auth.svelte";

    type Props = {
        showText?: boolean;
    };
    const { showText = true }: Props = $props();

	const session = useAuthSessionState();
    const preloadHome = $derived(session.error ? "tap" : "hover");
</script>

<div class="h-16 flex items-center justify-between px-4">
    <a href="/" class="text-2xl flex items-center" 
        data-sveltekit-preload-data={preloadHome} 
        data-sveltekit-preload-code={preloadHome}
    >
        <img src="/images/logo.svg" alt="logo" class="h-10 w-10 fill-neutral" />
        <span class={showText ? "pl-3" : "hidden"}>Rezible</span>
    </a>
</div>