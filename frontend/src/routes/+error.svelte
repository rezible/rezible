<script lang="ts">
	import { page } from "$app/state";
	import { Button } from "svelte-ux";
	import Card from "$components/card/Card.svelte";
	import Header from "$components/header/Header.svelte";

	const status = $derived(page.error?.status || "Unknown");
	const message = $derived(page.error?.message || "Error");

	console.log("error view", page.error);
</script>

<div class="w-96">
	<Card>
		{#snippet header()}
			<Header title="Error {status}" subheading={message} />
		{/snippet}
		<Button href="/">Go Home</Button>
	</Card>
</div>
