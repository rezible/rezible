<script lang="ts">
	import { Header } from "svelte-ux";
	import { collaboration } from '$features/incidents/lib/collaboration.svelte';
</script>

<div class="w-64 border overflow-y-auto p-2">
	<Header title="Context">
		<div slot="actions">
			<span>x</span>
		</div>
	</Header>
	<span>status: {collaboration.status}</span>
</div>