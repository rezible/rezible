<script lang="ts">
	import { page } from "$app/state";
	import OncallEventView from "$features/oncall-event/view/OncallEventView.svelte";
</script>

<OncallEventView id={page.params.id} />