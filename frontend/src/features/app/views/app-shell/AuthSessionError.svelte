<script lang="ts">
	import { session } from "$lib/auth.svelte";
	import HeaderLogo from "./HeaderLogo.svelte";

	console.log(session.error);
	const errorText = $derived(session.error ?? "unknown");
</script>

<div class="grid grid-rows-layout flex-1">
	<nav class="w-full h-16 border-b">
		<HeaderLogo expandingHover={false} />
	</nav>

	<main class="w-full overflow-y-auto p-2 grid place-items-center">
		<div class="border w-64">
			<span class="m-4 text-lg text-danger">
				auth error: {errorText}
			</span>
		</div>
	</main>
</div>
