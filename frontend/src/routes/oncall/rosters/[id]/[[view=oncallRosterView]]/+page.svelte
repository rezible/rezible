<script lang="ts">
	import { page } from "$app/state";
	import { convertOncallRosterViewParam } from "$src/params/oncallRosterView";
	import OncallRosterView from "$features/oncall/views/roster/OncallRosterView.svelte";

	const rosterId = $derived(page.params.id);
	const view = $derived(convertOncallRosterViewParam(page.params.view));
</script>

<OncallRosterView {rosterId} {view} />
