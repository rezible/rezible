<script lang="ts">
	import { type OncallRoster } from "$lib/api";
	import { appShell } from "$features/app/lib/appShellState.svelte";
	import PageActions from "./PageActions.svelte";
	import RosterStats from "./roster-stats/RosterStats.svelte";
	import RosterDetails from "./roster-details/RosterDetails.svelte";

	type Props = { roster: OncallRoster };
	const { roster }: Props = $props();

	appShell.setPageActions(PageActions, false);
</script>

<div class="flex gap-2 w-full h-full max-h-full min-h-0 overflow-hidden">
	<div class="w-1/3 grid grid-cols-2 grid-rows-3 gap-2">
		<RosterDetails {roster} />
	</div>

	<div class="w-2/3 border rounded-lg">
		<RosterStats {roster} />
	</div>
</div>
