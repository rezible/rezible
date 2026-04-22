<script lang="ts">
	import { useAppShell } from "$lib/app-shell.svelte";
	import * as Alert from "$components/ui/alert";
	import * as Card from "$components/ui/card";
	import { Button } from "$components/ui/button";
	import { Input } from "$components/ui/input";
	import { Label } from "$components/ui/label";
	import * as Select from "$components/ui/select";
	import { Separator } from "$components/ui/separator";
	import { Textarea } from "$components/ui/textarea";

	import { initOncallRosterEditorViewController } from "./controller.svelte";

	const { slug }: { slug?: string } = $props();

	const view = initOncallRosterEditorViewController(() => slug);

	const appShell = useAppShell();


	const selectedSchedule = $derived(view.selectedSchedule);
	const availableUsers = $derived(view.availableUsersForSelectedSchedule);
</script>

<div class="flex h-full flex-col gap-4">
	<Alert.Root>
		<Alert.Title>Draft editor wired to the current oncall roster query shape</Alert.Title>
		<Alert.Description class="space-y-2 text-xs leading-5">
			<p>
				Existing roster data is loaded from `getOncallRoster` and participants from `listUsers`. The
				fields below are still mock-backed in `mock_data.ts` because the read API does not expose them
				yet.
			</p>
			<ul class="list-disc pl-4">
				{#each view.mockBackedFields as field}
					<li><span class="font-medium">{field.label}:</span> {field.reason}</li>
				{/each}
			</ul>
		</Alert.Description>
	</Alert.Root>

	{#if view.loading}
		<Card.Root>
			<Card.Content class="py-8 text-sm text-surface-content/70">Loading roster draft…</Card.Content>
		</Card.Root>
	{:else}
		<div class="grid gap-4 xl:grid-cols-[minmax(0,2fr)_minmax(320px,1fr)]">
			<div class="flex min-w-0 flex-col gap-4">
				<Card.Root>
					<Card.Header>
						<Card.Title>{view.title}</Card.Title>
						<Card.Description>{view.subtitle}</Card.Description>
					</Card.Header>
					<Card.Content class="grid gap-4 md:grid-cols-2">
						<div class="flex flex-col gap-2">
							<Label for="roster-name">Roster Name</Label>
							<Input
								id="roster-name"
								value={view.draft.name}
								oninput={(event) => view.setRosterName(event.currentTarget.value)}
								placeholder="Platform Primary"
							/>
						</div>

						<div class="flex flex-col gap-2">
							<Label for="roster-slug">Roster Slug</Label>
							<Input
								id="roster-slug"
								value={view.draft.slug}
								oninput={(event) => view.setRosterSlug(event.currentTarget.value)}
								placeholder="platform-primary"
							/>
						</div>

						<div class="flex flex-col gap-2">
							<Label for="roster-timezone">Roster Timezone</Label>
							<Select.Root
								type="single"
								name="roster-timezone"
								bind:value={() => view.draft.timezone, view.setRosterTimezone}
							>
								<Select.Trigger class="w-full">
									{view.draft.timezone || "Select timezone"}
								</Select.Trigger>
								<Select.Content>
									{#each view.timezoneOptions as timezone}
										<Select.Item value={timezone} label={timezone}>{timezone}</Select.Item
										>
									{/each}
								</Select.Content>
							</Select.Root>
						</div>

						<div class="flex flex-col gap-2">
							<Label for="handover-template-id">Handover Template Id</Label>
							<Input
								id="handover-template-id"
								value={view.draft.handoverTemplateId}
								oninput={(event) =>
									view.setRosterHandoverTemplateId(event.currentTarget.value)}
								placeholder="uuid"
							/>
						</div>

						<div class="flex flex-col gap-2">
							<Label for="chat-handle">Chat Handle</Label>
							<Input
								id="chat-handle"
								value={view.draft.chatHandle}
								oninput={(event) => view.setRosterChatHandle(event.currentTarget.value)}
								placeholder="@platform-oncall"
							/>
						</div>

						<div class="flex flex-col gap-2">
							<Label for="chat-channel-id">Chat Channel Id</Label>
							<Input
								id="chat-channel-id"
								value={view.draft.chatChannelId}
								oninput={(event) => view.setRosterChatChannelId(event.currentTarget.value)}
								placeholder="slack:C07PLATFORM"
							/>
						</div>
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header class="gap-3">
						<div class="flex flex-wrap items-center justify-between gap-3">
							<div>
								<Card.Title>Schedules</Card.Title>
								<Card.Description>{view.scheduleCountLabel}</Card.Description>
							</div>
							<div class="flex gap-2">
								<Button variant="outline" onclick={() => view.addSchedule()}
									>Add Schedule</Button
								>
								<Button
									variant="destructive"
									disabled={view.draft.schedules.length <= 1}
									onclick={() => view.removeSelectedSchedule()}
								>
									Remove Selected
								</Button>
							</div>
						</div>
					</Card.Header>
					<Card.Content class="grid gap-4 lg:grid-cols-[240px_minmax(0,1fr)]">
						<div class="flex min-w-0 flex-col gap-2">
							{#each view.draft.schedules as schedule}
								<button
									type="button"
									class={`bg-background hover:bg-muted flex w-full flex-col items-start gap-1 border p-3 text-left text-xs transition-colors ${
										selectedSchedule?.key === schedule.key
											? "border-primary bg-primary/5"
											: "border-border"
									}`}
									onclick={() => view.selectSchedule(schedule.key)}
								>
									<span class="font-medium">{schedule.name || "Untitled Schedule"}</span>
									<span class="text-surface-content/70"
										>{schedule.timezone || "No timezone"}</span
									>
									<span class="text-surface-content/60">
										{schedule.participants.length} participant{schedule.participants
											.length === 1
											? ""
											: "s"}
									</span>
								</button>
							{/each}
						</div>

						{#if selectedSchedule}
							<div class="flex min-w-0 flex-col gap-4">
								<div class="grid gap-4 md:grid-cols-2">
									<div class="flex flex-col gap-2">
										<Label for="schedule-name">Schedule Name</Label>
										<Input
											id="schedule-name"
											value={selectedSchedule.name}
											oninput={(event) =>
												view.setSelectedScheduleName(event.currentTarget.value)}
											placeholder="Primary Rotation"
										/>
									</div>

									<div class="flex flex-col gap-2">
										<Label for="schedule-timezone">Schedule Timezone</Label>
										<Select.Root
											type="single"
											name="schedule-timezone"
											bind:value={
												() => selectedSchedule.timezone,
												view.setSelectedScheduleTimezone
											}
										>
											<Select.Trigger class="w-full">
												{selectedSchedule.timezone || "Select timezone"}
											</Select.Trigger>
											<Select.Content>
												{#each view.timezoneOptions as timezone}
													<Select.Item value={timezone} label={timezone}
														>{timezone}</Select.Item
													>
												{/each}
											</Select.Content>
										</Select.Root>
									</div>
								</div>

								<div class="flex flex-col gap-2">
									<Label for="schedule-description">Schedule Notes</Label>
									<Textarea
										id="schedule-description"
										rows={4}
										value={selectedSchedule.description}
										oninput={(event) =>
											view.setSelectedScheduleDescription(event.currentTarget.value)}
										placeholder="Shift goals, escalation notes, or handover context."
									/>
								</div>

								<Separator />

								<div class="flex flex-col gap-3">
									<div class="flex flex-wrap items-end gap-2">
										<div class="min-w-64 flex-1">
											<Label for="participant-picker">Add Participant</Label>
											<Select.Root
												type="single"
												name="participant-picker"
												bind:value={
													() => view.pendingParticipantUserId,
													(value) => (view.pendingParticipantUserId = value)
												}
												disabled={!availableUsers.length}
											>
												<Select.Trigger class="mt-2 w-full">
													{#if view.pendingParticipantUserId}
														{view.getUser(view.pendingParticipantUserId)
															?.attributes.name ?? "Select a user"}
													{:else}
														{availableUsers.length
															? "Select a user"
															: "All users already assigned"}
													{/if}
												</Select.Trigger>
												<Select.Content>
													{#each availableUsers as user}
														<Select.Item
															value={user.id}
															label={user.attributes.name}
														>
															{user.attributes.name}
														</Select.Item>
													{/each}
												</Select.Content>
											</Select.Root>
										</div>

										<Button
											variant="outline"
											disabled={!view.pendingParticipantUserId}
											onclick={() =>
												view.addSelectedScheduleParticipant(
													view.pendingParticipantUserId
												)}
										>
											Add To Rotation
										</Button>
									</div>

									<div class="flex flex-col gap-2">
										{#each selectedSchedule.participants as participant, index}
											{@const participantUser = view.getUser(participant.userId)}
											<div
												class="border-border flex items-center justify-between gap-3 border p-3 text-xs"
											>
												<div class="min-w-0">
													<div class="truncate font-medium">
														{participantUser?.attributes.name ??
															participant.userId}
													</div>
													<div class="text-surface-content/70">
														{participantUser?.attributes.email ??
															"User details loading"}
													</div>
												</div>

												<div class="flex gap-2">
													<Button
														size="sm"
														variant="outline"
														disabled={index === 0}
														onclick={() =>
															view.moveSelectedScheduleParticipant(
																participant.userId,
																-1
															)}
													>
														Up
													</Button>
													<Button
														size="sm"
														variant="outline"
														disabled={index ===
															selectedSchedule.participants.length - 1}
														onclick={() =>
															view.moveSelectedScheduleParticipant(
																participant.userId,
																1
															)}
													>
														Down
													</Button>
													<Button
														size="sm"
														variant="destructive"
														onclick={() =>
															view.removeSelectedScheduleParticipant(
																participant.userId
															)}
													>
														Remove
													</Button>
												</div>
											</div>
										{:else}
											<div
												class="border-border text-surface-content/70 border border-dashed p-4 text-xs"
											>
												No participants assigned to this schedule yet.
											</div>
										{/each}
									</div>
								</div>
							</div>
						{/if}
					</Card.Content>
				</Card.Root>
			</div>

			<div class="flex min-w-0 flex-col gap-4">
				<Card.Root>
					<Card.Header>
						<Card.Title>Draft Status</Card.Title>
						<Card.Description>State derived from the editor controller.</Card.Description>
					</Card.Header>
					<Card.Content class="space-y-3 text-xs">
						<div class="flex items-center justify-between gap-3">
							<span class="text-surface-content/70">Mode</span>
							<span class="font-medium capitalize">{view.mode}</span>
						</div>
						<div class="flex items-center justify-between gap-3">
							<span class="text-surface-content/70">Unsaved Changes</span>
							<span class="font-medium">{view.hasUnsavedChanges ? "Yes" : "No"}</span>
						</div>
						<div class="flex items-center justify-between gap-3">
							<span class="text-surface-content/70">Roster Id</span>
							<span class="font-medium">{view.rosterId ?? "New roster"}</span>
						</div>
						<div class="flex items-center justify-between gap-3">
							<span class="text-surface-content/70">Schedule Count</span>
							<span class="font-medium">{view.draft.schedules.length}</span>
						</div>
					</Card.Content>
					<Card.Footer class="flex justify-between gap-2">
						<Button
							variant="outline"
							href={view.mode === "edit" && view.rosterSlug
								? `/rosters/${view.rosterSlug}`
								: "/rosters"}
						>
							Cancel
						</Button>
						<Button disabled>{view.submitLabel}</Button>
					</Card.Footer>
				</Card.Root>

				<Card.Root>
					<Card.Header>
						<Card.Title>OncallRoster Schema Preview</Card.Title>
						<Card.Description>Draft fields mapped to the backend roster schema.</Card.Description>
					</Card.Header>
					<Card.Content>
						<pre class="bg-muted overflow-x-auto p-3 text-[11px] leading-5">{JSON.stringify(
								view.rosterSchemaPreview,
								null,
								2
							)}</pre>
					</Card.Content>
				</Card.Root>

				<Card.Root>
					<Card.Header>
						<Card.Title>OncallSchedule Schema Preview</Card.Title>
						<Card.Description
							>Draft schedules mapped to backend schedule and participant fields.</Card.Description
						>
					</Card.Header>
					<Card.Content>
						<pre class="bg-muted overflow-x-auto p-3 text-[11px] leading-5">{JSON.stringify(
								view.scheduleSchemaPreview,
								null,
								2
							)}</pre>
					</Card.Content>
				</Card.Root>
			</div>
		</div>
	{/if}
</div>
