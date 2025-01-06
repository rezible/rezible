<script lang="ts">
	import type { Incident } from "$lib/api";
    import SystemDiagram from "./SystemDiagram.svelte";

	type Props = {
		incident: Incident; 
	}
	const { incident }: Props = $props();
</script>

<div class="flex flex-col h-full w-full overflow-y-hidden">
	<div class="min-h-0 flex-1">
		<SystemDiagram />
	</div>
	<div class="h-20">
		actions
	</div>
</div>
