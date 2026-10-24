<script lang="ts">
	import { Button } from "$components/ui/button";
	import { Badge } from "$components/ui/badge";
	import { Textarea } from "$components/ui/textarea";
	import * as Dialog from "$components/ui/dialog";
	import * as Field from "$components/ui/field";
	import * as Avatar from "$components/ui/avatar";
	import StatusBadge from "$components/common/status-badge/StatusBadge.svelte";
	import MemberPicker from "$components/common/member-picker/MemberPicker.svelte";
	import InlineAlert from "$components/layout/error-alert/ErrorAlert.svelte";
	import { initReviewPanel } from "./controller.svelte";

	const review = initReviewPanel();
	const model = $derived(review.model);
</script>

<section class="flex flex-col gap-4" aria-label="Review">
	<div class="flex items-center justify-between gap-2">
		<h2 class="font-semibold">Review</h2>
		<StatusBadge status={model.status} />
	</div>
	<p class="text-sm text-muted-foreground">{model.explanation}</p>
	{#if review.error && !review.dialog}
		<InlineAlert error={review.error} dismissable={false} />
	{/if}
	<div class="flex items-center gap-2 text-sm">
		<span class="text-muted-foreground">Owner</span>
		{#if model.owner}
			<Avatar.Root class="size-6">
				<Avatar.Fallback>{model.owner.attributes?.name?.slice(0, 1) ?? "?"}</Avatar.Fallback>
			</Avatar.Root>
			<span>{model.owner.attributes?.name ?? "Member"}</span>
		{:else}
			<span>No owner</span>
		{/if}
		{#if model.actions.setOwner}
			<Button variant="link" size="sm" onclick={() => review.open("owner")}>Change</Button>
		{/if}
	</div>
	{#if model.actions.takeOwnership}
		<Button variant="outline" size="sm" onclick={review.take} disabled={review.pending}>
			Take ownership
		</Button>
	{/if}
	{#each model.reviewers as row (row.reviewer.id)}
		<div class="flex flex-col gap-2 text-sm">
			<div class="flex items-center gap-2">
				<Avatar.Root class="size-6">
					<Avatar.Fallback>{row.name.slice(0, 1)}</Avatar.Fallback>
				</Avatar.Root>
				<span class="min-w-0 flex-1">
					{row.name}{model.ownReview?.reviewer.id === row.reviewer.id ? " (you)" : ""}
				</span>
				<Badge variant="outline">{row.label}</Badge>
				{#if model.actions.changeReviewers}
					<Button
						variant="ghost"
						size="icon-sm"
						aria-label={`Remove ${row.name}`}
						disabled={review.pending}
						onclick={() => review.remove(row.reviewer.id, row.name)}
					>
						×
					</Button>
				{/if}
			</div>
			{#if row.feedback}
				<details>
					<summary class="cursor-pointer text-muted-foreground">Feedback · Show more</summary>
					<blockquote class="border-l-2 pl-3 whitespace-pre-wrap">{row.feedback}</blockquote>
				</details>
			{/if}
			{#if row.messageFailed && model.isOwner}
				<p class="text-xs text-muted-foreground">Couldn't send a message</p>
			{/if}
		</div>
	{/each}
	{#if model.actions.changeReviewers}
		<Button variant="link" size="sm" onclick={() => review.open("add")}>Add reviewer</Button>
	{/if}
	{#if model.actions.requestReview}
		<Button size="sm" onclick={() => review.open("request")} disabled={review.pending}>
			Request review
		</Button>
	{/if}
	{#if model.actions.close}
		<span title={model.closeReason}>
			<Button
				variant="outline"
				size="sm"
				onclick={() => review.open("close")}
				disabled={!model.actions.canClose || review.pending}
			>
				{model.closeLabel}
			</Button>
		</span>
	{/if}
	{#if model.actions.skip}
		<Button variant="ghost" size="sm" onclick={() => review.open("skip")} disabled={review.pending}>
			Skip
		</Button>
	{/if}
	{#if model.actions.decide}
		<div class="flex flex-wrap gap-2">
			<Button size="sm" onclick={review.approve} disabled={review.pending}>Approve</Button>
			<Button variant="outline" size="sm" onclick={review.requestChanges} disabled={review.pending}>
				Request changes
			</Button>
		</div>
		{#if model.ownReview?.state !== "waiting"}
			<p class="text-xs text-muted-foreground">Change your decision while the review is open.</p>
		{/if}
		{#if review.showFeedback}
			<Field.Field>
				<Field.FieldLabel for="review-feedback">What needs to change?</Field.FieldLabel>
				<Textarea id="review-feedback" bind:value={review.feedback} maxlength={4000} />
				<Button
					size="sm"
					onclick={review.sendFeedback}
					disabled={!review.feedback.trim() || review.pending}
				>
					Send
				</Button>
			</Field.Field>
		{/if}
	{/if}
	{#if model.closed && review.skipReason}
		<blockquote class="border-l-2 pl-3 text-sm whitespace-pre-wrap">{review.skipReason}</blockquote>
	{/if}
</section>

<Dialog.Root
	open={!!review.dialog}
	onOpenChange={(open) => {
		if (!open) {
			review.cancel();
		}
	}}
>
	<Dialog.Content>
		<Dialog.Header>
			<Dialog.Title>{review.dialogTitle}</Dialog.Title>
			<Dialog.Description>
				{#if review.dialog === "close"}
					{model.closeLabel === "Close"
						? "Everyone has approved. This can't be undone."
						: "The retrospective will be marked closed without review. This can't be undone."}
				{:else if review.dialog === "skip"}
					Shown with the incident.
				{:else if review.dialog === "request" || review.dialog === "add"}
					Each reviewer is messaged in chat if Rezible knows their account.
				{:else}
					Confirm this change to the retrospective.
				{/if}
			</Dialog.Description>
		</Dialog.Header>
		{#if review.error}
			<InlineAlert error={review.error} dismissable={false} />
		{/if}
		<Field.FieldGroup>
			{#if review.dialog === "owner" || review.dialog === "request" || review.dialog === "add"}
				<Field.Field>
					<Field.FieldLabel>Members</Field.FieldLabel>
					<MemberPicker
						selectedIds={review.selectedIds}
						excludeIds={review.excludeIds}
						multiple={review.dialog === "request"}
						onSelect={review.select}
					>
						{#snippet trigger(props)}
							<Button {...props} variant="outline">
								{review.selectedNames || "Choose members"}
							</Button>
						{/snippet}
					</MemberPicker>
				</Field.Field>
			{:else if review.dialog === "skip"}
				<Field.Field>
					<Field.FieldLabel for="skip-reason">Why is no retrospective needed?</Field.FieldLabel>
					<Textarea
						id="skip-reason"
						bind:value={review.reason}
						maxlength={4000}
						placeholder="Minor config error, already covered by INC-131."
					/>
				</Field.Field>
			{/if}
		</Field.FieldGroup>
		<Dialog.Footer>
			<Button variant="ghost" onclick={review.cancel} disabled={review.pending}>Cancel</Button>
			<Button
				variant={review.dialog === "skip" || review.dialog === "remove" ? "destructive" : "default"}
				onclick={review.submit}
				disabled={!review.canSubmit}
			>
				{review.pending ? "Saving…" : "Confirm"}
			</Button>
		</Dialog.Footer>
	</Dialog.Content>
</Dialog.Root>
