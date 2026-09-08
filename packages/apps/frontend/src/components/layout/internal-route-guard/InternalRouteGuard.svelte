<script lang="ts">
	import { goto } from "$app/navigation";
	import type { Snippet } from "svelte";
	import { useUserSessionState } from "$lib/user-session.svelte";
	import { Spinner } from "$components/ui/spinner";

	type Props = {
		children: Snippet;
	};
	const { children }: Props = $props();

	const session = useUserSessionState();

	const authorized = $derived(session.ready && session.isAdmin);

	$effect(() => {
		if (session.ready && !session.isAdmin) {
			goto("/");
		}
	});
</script>

{#if authorized}
	{@render children()}
{:else}
	<div class="grid flex-1 place-items-center">
		<Spinner />
	</div>
{/if}
