<script lang="ts">
	import Avatar from "svelte-boring-avatars";

	type Props = {
		kind: "team" | "user" | "roster" | "service";
		id: string;
		size?: number;
		colors?: string[];
	};

	let {
		kind,
		id,
		size = 38,
		colors = ["#92A1C6", "#146A7C", "#F0AB3D", "#C271B4", "#C20D90"],
	}: Props = $props();

	const variants: Record<Props["kind"], Avatar["$$prop_def"]["variant"]> = {
		team: "bauhaus",
		user: "marble",
		roster: "sunset",
		service: "pixel",
	};
	const variant = variants[kind];
</script>

{#key id}
	<Avatar {variant} {size} name={id} {colors} />
{/key}
