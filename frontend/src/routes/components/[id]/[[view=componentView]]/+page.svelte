<script lang="ts">
	import { page } from "$app/state";
	import { SystemComponentView, setSystemComponentViewState } from "$features/system-component";

	setSystemComponentViewState(() => (page.params.id || ""));
</script>

<SystemComponentView />