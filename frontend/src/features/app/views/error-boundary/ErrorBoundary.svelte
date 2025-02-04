<script lang="ts">
	import { page } from "$app/state";
	import { Button, Card } from "svelte-ux";

	const status = $derived(page.error?.status || "Unknown");
	const message = $derived(page.error?.message || "Error");

	console.log("error view", page.error);
</script>

<div class="w-96">
	<Card title="Error {status}" subheading={message}>
		<Button href="/">Go Home</Button>
	</Card>
</div>
