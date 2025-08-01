<script lang="ts">
	import { useSystemComponentViewState } from "$features/system-component";

	const view = useSystemComponentViewState();
</script>

<span>name: {view.componentName}</span>