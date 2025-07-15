<script lang="ts">
	import { page } from "$app/state";
	import AlertView from "$features/alerts/views/alert/AlertView.svelte";

	const id = $derived(page.params.id);
</script>

<AlertView {id} />