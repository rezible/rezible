<script lang="ts">
	import { AUTH_REDIRECT_URL, session } from "$lib/auth.svelte";
	import { Button } from "svelte-ux";

	const errorText = $derived(session.error?.category ?? "unknown");
	
	// clear the session if user is not found
	const authPath = $derived(session.error?.category === "no_user" ? "/logout" : "");
	const buttonHref = $derived(`${AUTH_REDIRECT_URL}${authPath}`);
</script>

<div class="grid place-items-center">
    <div class="w-64 flex flex-col gap-2">
        {#if session.error}
            <span class="text-danger">Auth Error: {errorText}</span>
        {/if}

        <Button href={buttonHref} color="primary" variant="fill">Sign in</Button>
    </div>
</div>