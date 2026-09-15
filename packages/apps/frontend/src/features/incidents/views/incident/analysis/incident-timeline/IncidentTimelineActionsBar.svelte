<script lang="ts">
	import { Button } from "$components/ui/button";
	import { useEventDialog } from "./event-dialog/controller.svelte";

	const editor = useEventDialog();
</script>

<Button variant="outline" size="sm" onclick={() => editor.setCreating()}>Create entry</Button>
