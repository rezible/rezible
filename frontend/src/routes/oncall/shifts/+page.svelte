<script>
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import ShiftsList from "$features/oncall/views/shifts-list/ShiftsList.svelte";

	appShell.setPageBreadcrumbs(() => [{ label: "Oncall", href: "/oncall" }, { label: "Shifts" }])
</script>

<ShiftsList />
