<script lang="ts">
	import { resolve } from "$app/paths";
	import { Button } from "$components/ui/button";
	import * as Card from "$components/ui/card";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";

	registerPageDescriptor(() => ({
		title: "Access denied",
		parents: [{ label: "Preferences", path: resolve("/settings/user") }],
	}));
</script>

<div class="flex min-h-full min-w-0 items-start justify-center pt-8">
	<Card.Root size="sm" class="w-full max-w-md">
		<Card.Header>
			<Card.Title>Organization settings</Card.Title>
			<Card.Description>You need administrator access to view this page.</Card.Description>
		</Card.Header>
		<Card.Footer>
			<Button href={resolve("/settings/user")}>Back to preferences</Button>
		</Card.Footer>
	</Card.Root>
</div>
