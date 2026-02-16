<script lang="ts">
	import { page } from "$app/state";
	import SystemComponentView from "$features/system/views/system-component-view";
</script>

<SystemComponentView id={page.params.id || ""} />