<script lang="ts">
	import { page } from "$app/state";
	import PlaybookView from "$features/playbooks/views/playbook";
</script>

<PlaybookView id={page.params.id || ""} />