<script lang="ts">
	import { page } from "$app/state";
	import { OncallShiftView, setOncallShiftViewState } from "$features/oncall-shift";

	setOncallShiftViewState(() => page.params.id);
</script>

<OncallShiftView />
