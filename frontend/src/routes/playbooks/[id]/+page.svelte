<script lang="ts">
	import { page } from "$app/state";
	import PlaybookView from "$features/playbooks/views/playbook/PlaybookView.svelte";
</script>

<PlaybookView id={page.params.id} />