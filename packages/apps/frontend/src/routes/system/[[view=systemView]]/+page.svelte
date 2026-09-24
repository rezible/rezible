<script lang="ts">
	import type { PageProps } from "./$types";
	import SystemView from "$features/system/views/system";
	import { isSystemViewParam } from "$params/systemView";

	type Props = PageProps;
	let { params }: Props = $props();
	
	const view = $derived(isSystemViewParam(params.view) ? params.view : undefined);
</script>

<SystemView {view} />
