<script lang="ts">
	import { Button, TextField } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiLink, mdiPlus } from "@mdi/js";

	type Props = {
		onLinked: (id: string) => void;
	};
	const { onLinked }: Props = $props();
</script>

<TextField label="Link New Item" classes={{ input: "text-sm" }}>
	<div slot="prepend">
		<Icon data={mdiLink} classes={{root: "text-surface-content/50 mr-2"}} />
	</div>
	<span slot="append">
		<Button icon={mdiPlus} class="text-surface-content/50 p-2" />
	</span>
</TextField>
