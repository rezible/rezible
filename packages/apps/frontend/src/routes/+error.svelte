<script lang="ts">
	import { page } from "$app/state";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";

	const status = $derived(page.error?.status || "Unknown");
	const message = $derived(page.error?.message || "Error");

	console.log("error view", page.error);
</script>

<div class="w-full h-full grid place-items-center">

	<Card.Root class="gap-4 p-4 min-w-xs w-96">
		<Card.Header class="p-0">
            <Card.Title class="capitalize">Error {status}</Card.Title>
			<Card.Description>{message}</Card.Description>
		</Card.Header>
		<Card.Action>
			<Button href="/">Go Home</Button>
		</Card.Action>
	</Card.Root>
</div>