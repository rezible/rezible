<script>
	import IncidentFields from "./IncidentFields.svelte";
	import IncidentRoles from "./IncidentRoles.svelte";
	import IncidentTypes from "./IncidentTypes.svelte";
	import IncidentTags from "./IncidentTags.svelte";
</script>

<div class="flex flex-col gap-2 max-h-full overflow-y-auto">
	<IncidentFields />

	<IncidentRoles />

	<IncidentTypes />

	<IncidentTags />
</div>
