<script lang="ts">
	import * as Card from "$components/ui/card";
	import { resolve } from "$app/paths";
	import { registerPageDescriptor } from "$lib/app-shell.svelte";
	import { initOrganizationSettingsViewController } from "./organizationSettingsController.svelte";

	const view = initOrganizationSettingsViewController();

	registerPageDescriptor(() => ({
		title: "Organization",
		parents: [{ label: "Settings", path: resolve("/settings") }],
	}));
</script>

<Card.Root class="max-w-2xl">
	<Card.Header>
		<Card.Title>Organization</Card.Title>
	</Card.Header>
	<Card.Content>
		<div class="grid gap-1">
			<div class="text-sm font-medium text-muted-foreground">Name</div>
			<div>{view.orgName || "Organization"}</div>
		</div>
	</Card.Content>
</Card.Root>
