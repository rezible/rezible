<script lang="ts">
	import { Button, Drawer } from "svelte-ux";


	type Props = {
		target: HTMLElement;
		open: boolean;
	}
	let { target, open = $bindable() }: Props = $props();
	
	const setClosed = () => {
		open = false;
	}
</script>

<Drawer
    {open}
    on:close={setClosed}
    portal={{ target: "ctx-sidebar" }}
  >
    <h1>Contents</h1>
    <div slot="actions">
      <Button on:click={setClosed}>Close</Button>
    </div>
</Drawer>