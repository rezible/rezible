<script lang="ts">
	import { useAuthSessionState } from "$lib/auth.svelte";
	import { onMount } from "svelte";
	import { url } from "zod";

	const auth = useAuthSessionState();

	onMount(() => {
		auth.onSignInCallback();
	});
</script>

<span>loading: {auth.isAuthenticated}</span>
