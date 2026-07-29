import {
	getUserSessionOptions,
	getUserSessionPreferencesOptions,
	type ErrorModel,
	type UserSessionNotificationPreferences,
	updateUserSessionPreferencesMutation,
} from "$lib/api";
import { useUserSessionState } from "$lib/user-session.svelte";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";

type NotificationKey = keyof UserSessionNotificationPreferences;

const defaultNotifications: UserSessionNotificationPreferences = {
	incidentUpdates: true,
	incidentRoleAssignments: true,
	agentRunResults: true,
	integrationSyncFailures: true,
};

export class UserSettingsController {
	private session = useUserSessionState();
	private queryClient = useQueryClient();

	private preferencesQuery = createQuery(() => getUserSessionPreferencesOptions());
	preferences = $derived(this.preferencesQuery.data?.data);
	loading = $derived(this.preferencesQuery.isPending);
	error = $derived(this.preferencesQuery.error as ErrorModel | null);
	saveError = $state<ErrorModel>();

	name = $state("");
	email = $state("");
	timezone = $state("");
	notifications = $state<UserSessionNotificationPreferences>({ ...defaultNotifications });

	private updatePreferencesMut = createMutation(() => ({
		...updateUserSessionPreferencesMutation(),
		onSuccess: async () => {
			this.saveError = undefined;
			await this.queryClient.invalidateQueries({
				queryKey: getUserSessionPreferencesOptions().queryKey,
			});
			await this.queryClient.invalidateQueries({ queryKey: getUserSessionOptions().queryKey });
			this.session.refetch();
		},
		onError: (err) => {
			this.saveError = err;
		},
	}));
	saving = $derived(this.updatePreferencesMut.isPending);

	constructor() {
		watch(
			() => this.preferences,
			(preferences) => {
				if (!preferences) return;
				this.name = preferences.profile.name;
				this.email = preferences.profile.email;
				this.timezone = preferences.profile.timezone;
				this.notifications = { ...defaultNotifications, ...preferences.notifications };
			}
		);
	}

	saveProfile() {
		this.updatePreferencesMut.mutate({
			body: { attributes: { name: this.name.trim(), timezone: this.timezone.trim() } },
		});
	}

	saveNotifications() {
		this.updatePreferencesMut.mutate({
			body: { attributes: { ...this.notifications } },
		});
	}

	setNotification(key: NotificationKey, value: boolean) {
		this.notifications = { ...this.notifications, [key]: value };
	}
}

const ctx = new Context<UserSettingsController>("UserSettingsController");
export const initUserSettingsController = () => ctx.set(new UserSettingsController());
