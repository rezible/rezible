<script lang="ts">
	import { page } from "$app/state";
	import SystemComponentView from "$features/system-components/views/component/SystemComponentView.svelte";
</script>

<SystemComponentView id={page.params.id} />