<script lang="ts">
	import Toaster, { setToastState } from "$components/toaster";
	import Sidebar from "./Sidebar.svelte";
	import Header from "./Header.svelte";

	const { children } = $props();

	setToastState();
</script>

<div class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-200 text-surface-content">
	<Sidebar />

	<div class="grid grid-rows-layout flex-1">
		<nav class="w-full h-16">
			<Header />
		</nav>

		<main class="w-full overflow-y-auto px-2 pb-2">
			{@render children()}
		</main>
	</div>
</div>

<Toaster />
