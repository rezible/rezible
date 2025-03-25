<script lang="ts">
	import { page } from "$app/state";
	import { convertOncallShiftViewParam } from "$src/params/oncallShiftView";
	import OncallShiftOmniView from "$features/oncall/views/shift/OncallShiftOmniView.svelte";

	const shiftId = $derived(page.params.id);

	const view = $derived(convertOncallShiftViewParam(page.params.view));
</script>

<OncallShiftOmniView {shiftId} {view} />
