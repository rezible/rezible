<script lang="ts">
	import type { OncallShift } from "$lib/api";
	import type { ShiftEvent } from "$features/oncall/lib/utils";

	type Props = {
		shift: OncallShift;
		shiftEvents: ShiftEvent[];
	};
	let { shift, shiftEvents }: Props = $props();

</script>