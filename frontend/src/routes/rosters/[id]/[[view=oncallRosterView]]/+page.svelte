<script lang="ts">
	import { page } from "$app/state";
	import { convertOncallRosterViewParam } from "$src/params/oncallRosterView";
	import OncallRosterView from "$features/oncall-roster/view/OncallRosterView.svelte";
</script>

<OncallRosterView 
	id={page.params.id}
	view={convertOncallRosterViewParam(page.params.view)}
/>
