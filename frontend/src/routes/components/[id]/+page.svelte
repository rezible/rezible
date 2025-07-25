<script lang="ts">
	import { page } from "$app/state";
	import SystemComponentView from "$src/features/system-component/view/SystemComponentView.svelte";
</script>

<SystemComponentView id={page.params.id} />