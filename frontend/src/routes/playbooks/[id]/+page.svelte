<script lang="ts">
	import { page } from "$app/state";
	import PlaybookView from "$features/playbooks/views/playbook/PlaybookView.svelte";

	const id = $derived(page.params.id);
</script>

<PlaybookView {id} />