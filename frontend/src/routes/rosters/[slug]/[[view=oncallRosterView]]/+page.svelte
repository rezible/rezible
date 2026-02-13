<script lang="ts">
	import { page } from "$app/state";
	import { OncallRosterView, setOncallRosterViewState } from "$features/oncall-roster";

	setOncallRosterViewState(() => (page.params.slug || ""));
</script>

<OncallRosterView />
