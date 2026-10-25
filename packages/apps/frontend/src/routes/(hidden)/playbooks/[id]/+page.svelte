<script lang="ts">
	import type { PageProps } from "./$types";
	import PlaybookView from "$features/playbooks/views/playbook";

	const { params }: PageProps = $props();
</script>

<PlaybookView id={params.id} />
