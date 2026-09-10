<script lang="ts">
	import { Button } from "$components/ui/button";
	import RiLink from "remixicon-svelte/icons/link";
	import RiAddLine from "remixicon-svelte/icons/add-line";

	type Props = {
		onLinked: (id: string) => void;
	};
	const { onLinked }: Props = $props();
</script>

<!-- <TextField label="Link New Item" classes={{ input: "text-sm" }}>
	<div slot="prepend">
		<RiLink class="text-foreground/50 mr-2" aria-hidden="true" />
	</div>
	<span slot="append">
		<Button class="text-foreground/50 p-2" />
	</span>
</TextField> -->
