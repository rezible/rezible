<script lang="ts">
	import { page } from "$app/state";
	import PlaybookView from "$src/features/playbook/view/PlaybookView.svelte";
</script>

<PlaybookView id={page.params.id} />