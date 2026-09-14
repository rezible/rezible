<script lang="ts">
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import SystemMapCanvas from "./SystemMapCanvas.svelte";

	registerPageDescriptor(() => ({ title: "System Map" }));
</script>

<SystemMapCanvas />
