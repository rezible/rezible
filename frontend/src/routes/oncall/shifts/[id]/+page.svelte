<script lang="ts">
	import { page } from "$app/state";
	import OncallShiftView from "$src/features/oncall/views/shift/OncallShiftView.svelte";
</script>

<OncallShiftView shiftId={page.params.id} />
