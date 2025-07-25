<script lang="ts">
	import { page } from "$app/state";
	import AlertView from "$features/alert/view/AlertView.svelte";
</script>

<AlertView id={page.params.id} />