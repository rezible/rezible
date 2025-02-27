<script lang="ts">
	import { Button, Header } from "svelte-ux";
	import { mdiChevronRight } from "@mdi/js";
	import { collaboration } from "$features/incidents/lib/collaboration.svelte";
</script>

<div class="w-64 border overflow-y-auto p-2">
	<Header title="Context">
		<div slot="actions">
			<Button iconOnly icon={mdiChevronRight} />
		</div>
	</Header>
	<span>status: {collaboration.status}</span>
</div>
