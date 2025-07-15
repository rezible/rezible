<script lang="ts">
	import { page } from "$app/state";
	import PlaybookView from "$features/playbooks/views/playbook/PlaybookView.svelte";

	const slug = $derived(page.params.slug);
</script>

<PlaybookView {slug} />