<script lang="ts">
	import { Button } from "svelte-ux";
	import Icon from "$components/icon/Icon.svelte";
	import { mdiChevronLeft, mdiChevronRight } from "@mdi/js";
	import { page } from "$app/state";

	const shiftId = $derived(page.params.id);
	// const currentView = $derived(page.params.view);
	// const query = createQuery(() => getOncallShiftOptions({ path: { id: shiftId } }));
	// const shift = $derived(query.data?.data);

	const previousShiftId = $derived(shiftId);
	const nextShiftId = $derived(shiftId);

	// const isActiveShift = $derived(shift ? buildShiftTimeDetails(shift).status === "active" : false);
	// const isSessionUser = $derived(session.userId == shift?.attributes.user.id);
</script>

<div class="flex gap-2">
	<Button href="/shifts/{previousShiftId}">
		<Icon data={mdiChevronLeft} />
		<span class="">Previous Shift</span>
	</Button>

	<!-- {#if shift && isActiveShift && isSessionUser}
		<Button
			href="/shifts/{shift.id}/handover"
			variant="fill"
			color="secondary"
		>
			<span>Edit Handover</span>
		</Button>
	{/if} -->

	<Button href="/shifts/{nextShiftId}">
		<span class="">Next Shift</span>
		<Icon data={mdiChevronRight} />
	</Button>
</div>