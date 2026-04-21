<script lang="ts">
	import { page } from "$app/state";
	import OncallShiftView from "$features/oncall/views/shift";
</script>

<OncallShiftView id={page.params.id || ""} />
