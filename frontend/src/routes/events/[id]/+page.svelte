<script lang="ts">
	import { page } from "$app/state";
	import { OncallEventView, setOncallEventViewState } from "$features/oncall-event";

	setOncallEventViewState(() => page.params.id);
</script>

<OncallEventView />