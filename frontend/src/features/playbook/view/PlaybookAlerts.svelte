<script lang="ts">
	import { usePlaybookViewState } from "$features/playbook";
	import { listAlertsOptions } from "$src/lib/api";
	import { createQuery } from "@tanstack/svelte-query";

	const view = usePlaybookViewState();
	
	// TODO: filters etc
	const query = createQuery(() => listAlertsOptions({query: {}}));
</script>