<script lang="ts">
	import { usePlaybookViewController } from "./controller.svelte"
	import { listAlertsOptions } from "$src/lib/api";
	import { createQuery } from "@tanstack/svelte-query";

	const view = usePlaybookViewController();
	
	// TODO: filters etc
	const query = createQuery(() => listAlertsOptions({query: {}}));
</script>