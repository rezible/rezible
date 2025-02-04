<script lang="ts">
	import { session } from "$lib/auth.svelte";

	console.log(session.error);
	const errorText = $derived(session.error ?? "unknown");
</script>

<div
	class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-300 text-surface-content"
>
	<span class="m-4 text-lg text-danger">
		auth error: {errorText}
	</span>
</div>
