<script lang="ts">
	import { mdiCircleMedium } from "@mdi/js";
	import Icon from "$components/icon/Icon.svelte";
	import { Notification } from "svelte-ux";
	import { getToastState, type Toast } from "$features/app-shell/lib/toasts.svelte";

	const toastState = getToastState();
</script>

{#snippet toast(t: Toast)}
	<Notification open closeIcon on:close={() => toastState.remove(t.id)}>
		<div slot="icon">
			<Icon data={t.icon ?? mdiCircleMedium} classes={{root: "text-success-500"}} />
		</div>
		<div slot="title">{t.title}</div>
		<div slot="description" class="w-64">{t.message}</div>
	</Notification>
{/snippet}

<div class="absolute right-4 bottom-4 flex flex-col gap-2 overflow-hidden">
	{#each toastState.toasts as t (t.id)}
		<div class="w-96">
			{@render toast(t)}
		</div>
	{/each}
</div>
