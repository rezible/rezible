<script lang="ts">
	import { AUTH_REDIRECT_URL, session } from "$lib/auth.svelte";
	import { Button } from "svelte-ux";

	const errorText = $derived(session.error?.category ?? "unknown");
</script>

<div class="grid place-items-center">
    <div class="w-64 flex flex-col gap-2">
        <span>auth</span>

        {#if session.error}
            <span class="text-danger">error: {errorText}</span>
        {/if}

        <Button href={AUTH_REDIRECT_URL} color="primary" variant="fill">Sign in</Button>
    </div>
</div>