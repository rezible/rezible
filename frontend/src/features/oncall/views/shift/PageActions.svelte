<script lang="ts">
	import { Button, Icon } from "svelte-ux";
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

	const id1 = "3c3c74a4-82d7-43a7-8e90-805e7a8f7d32";
	const id2 = "9a64f439-2349-49cf-91ff-997e69abee74";
	const otherShiftId = $derived(shiftId === id1 ? id2 : id1);
</script>

<div class="flex gap-2">
	<Button href="/oncall/shifts/{otherShiftId}">Other Shift</Button>

	<Button href="/oncall/shifts/{previousShiftId}">
		<Icon data={mdiChevronLeft} />
		<span class="">Previous Shift</span>
	</Button>

	<!-- {#if shift && isActiveShift && isSessionUser}
		<Button
			href="/oncall/shifts/{shift.id}/handover"
			variant="fill"
			color="secondary"
		>
			<span>Edit Handover</span>
		</Button>
	{/if} -->

	<Button href="/oncall/shifts/{nextShiftId}">
		<span class="">Next Shift</span>
		<Icon data={mdiChevronRight} />
	</Button>
</div>