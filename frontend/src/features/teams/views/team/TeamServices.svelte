<script lang="ts">
	import type { Service } from "$lib/api";
	import { AvatarRing } from "svelte-boring-avatars";

	interface Props {
		services: Service[];
	}
	let { services }: Props = $props();
</script>

<div class="flex flex-col gap-2 w-fit">
	{#each services as svc (svc.id)}
		<div class="flex items-center gap-2 px-2">
			<AvatarRing name={svc.id} size={32} />
			<span>{svc.attributes.name}</span>
		</div>
	{/each}
</div>
