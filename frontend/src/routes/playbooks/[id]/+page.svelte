<script lang="ts">
	import { page } from "$app/state";
	import { PlaybookView, setPlaybookViewState } from "$features/playbook";
	
	setPlaybookViewState(() => page.params.id);
</script>

<PlaybookView />