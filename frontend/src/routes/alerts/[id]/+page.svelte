<script lang="ts">
	import { page } from "$app/state";
	import AlertView from "$features/alerts/views/alert/AlertView.svelte";
</script>

<AlertView id={page.params.id} />