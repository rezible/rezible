<script lang="ts">
	import { page } from "$app/state";
	import ShiftHandoverView from "$features/oncall/views/shift-handover/ShiftHandoverView.svelte";
</script>

<ShiftHandoverView shiftId={page.params.id} />