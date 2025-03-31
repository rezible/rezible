<script lang="ts">
	import { AUTH_REDIRECT_URL, session } from "$lib/auth.svelte";
	import { Button, Header } from "svelte-ux";

	const errorText = $derived(session.error?.category ?? "unknown");
	
	// clear the session if user is not found
	const authPath = $derived(session.error?.category === "no_user" ? "/logout" : "");
	const buttonHref = $derived(`${AUTH_REDIRECT_URL}${authPath}`);
</script>

<div class="grid h-full w-full place-items-center">
    <div class="w-64 flex flex-col gap-2 border rounded-lg border-surface-content/10 bg-surface-200 p-3">
		<Header title="Authentication Required" class="border-b" />

        {#if session.error}
			<div class="bg-danger-900/50 border-danger/20 border rounded p-2">
				<span class="">{errorText}</span>
			</div>
        {/if}

        <Button href={buttonHref} color="primary" variant="fill">Sign in</Button>
    </div>
</div>