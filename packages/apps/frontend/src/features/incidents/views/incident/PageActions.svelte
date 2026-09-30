<script lang="ts">
	import { Button } from "$components/ui/button";
	import type { IncidentViewController } from "./controller.svelte";

	type Props = { controller: IncidentViewController };
	const { controller }: Props = $props();
</script>

<Button variant="outline" size="sm" onclick={controller.share}>Share</Button>
