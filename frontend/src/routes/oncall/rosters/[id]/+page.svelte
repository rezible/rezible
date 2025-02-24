<script lang="ts">
	import { page } from "$app/state";
	import OncallRosterView from "$features/oncall/views/roster/OncallRosterView.svelte";
</script>

<OncallRosterView rosterId={page.params.id} />