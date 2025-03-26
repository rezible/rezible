<script lang="ts">
	import { page } from "$app/state";
	import { convertOncallShiftViewParam } from "$src/params/oncallShiftView";
	import OncallShiftView from "$features/oncall/views/shift/OncallShiftView.svelte";

	const shiftId = $derived(page.params.id);

	const view = $derived(convertOncallShiftViewParam(page.params.view));
</script>

<OncallShiftView {shiftId} {view} />
