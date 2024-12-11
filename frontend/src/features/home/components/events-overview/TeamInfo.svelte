<script lang="ts">
    import { Header } from "svelte-ux";
    import TeamSelector from "./TeamSelector.svelte";

</script>

<div class="flex flex-col h-fit gap-2 border p-2 rounded-lg overflow-y-auto">
	<Header title="Team Information" class="border-b" />
	
	<TeamSelector onChange={t => {console.log("team:", t)}} />
</div>