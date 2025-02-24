<script lang="ts">
	import Toaster from "$components/toaster/Toaster.svelte";
	import Sidebar from "./Sidebar.svelte";
	import Header from "./Header.svelte";
	import PageContainer from "./PageContainer.svelte";

	const { children } = $props();
</script>

<div class="antialiased flex h-dvh min-h-dvh w-dvw bg-surface-200 text-surface-content">
	<Sidebar />

	<div class="grid grid-rows-layout flex-1">
		<nav class="w-full h-16">
			<Header />
		</nav>

		<main class="w-full overflow-y-auto px-2 pb-2">
			<PageContainer>
				{@render children()}
			</PageContainer>
		</main>
	</div>
</div>

<Toaster />
