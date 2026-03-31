<script lang="ts">
	import Avatar from "svelte-boring-avatars";

	type Props = {
		kind: "team" | "user" | "roster" | "service" | "incident";
		id: string;
		size?: number;
		square?: boolean;
		colors?: string[];
	};

	let {
		kind,
		id,
		size = 38,
		square = false,
		colors = ["#92A1C6", "#146A7C", "#F0AB3D", "#C271B4", "#C20D90"],
	}: Props = $props();

	const variants: Record<Props["kind"], Avatar["$$prop_def"]["variant"]> = {
		team: "bauhaus",
		user: "marble",
		roster: "sunset",
		service: "pixel",
		incident: "bauhaus",
	};
	const variant = variants[kind];
</script>

{#key id}
	<Avatar name={id} {variant} {square} {size} {colors} />
{/key}
