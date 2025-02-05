<script lang="ts" module>
	export type AvatarKind = "team" | "user" | "roster" | "service";
	export type AvatarProps = {
		kind: AvatarKind;
		id: string;
		size?: number;
		badge?: boolean;
		colors?: string[];
	};
</script>

<script lang="ts">
	import Avatar from "svelte-boring-avatars";
	import { Badge } from "svelte-ux";

	let {
		kind,
		id,
		size = 38,
		badge = false,
		colors = ["#92A1C6", "#146A7C", "#F0AB3D", "#C271B4", "#C20D90"],
	}: AvatarProps = $props();

	const variants: Record<AvatarKind, Avatar["$$prop_def"]["variant"]> = {
		team: "bauhaus",
		user: "marble",
		roster: "sunset",
		service: "pixel",
	};
	const variant = variants[kind];
</script>

{#snippet avatar()}
	<Avatar {variant} {size} name={id} {colors} />
{/snippet}

{#key id}
	{#if badge}
		<Badge placement="top-right" value={1} dot circle class="bg-success-200">
			{@render avatar()}
		</Badge>
	{:else}
		{@render avatar()}
	{/if}
{/key}
