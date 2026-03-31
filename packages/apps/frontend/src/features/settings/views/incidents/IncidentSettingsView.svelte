<script>
	import { appShell } from "$features/app";
	import IncidentFields from "./IncidentFields.svelte";
	import IncidentRoles from "./IncidentRoles.svelte";
	import IncidentTypes from "./IncidentTypes.svelte";
	import IncidentTags from "./IncidentTags.svelte";

	appShell.setPageBreadcrumbs(() => [
		{ label: "Settings", href: "/settings" },
		{ label: "Incidents" }
	]);
</script>

<div class="flex flex-col gap-2 max-h-full overflow-y-auto">
	<IncidentFields />

	<IncidentRoles />

	<IncidentTypes />

	<IncidentTags />
</div>
