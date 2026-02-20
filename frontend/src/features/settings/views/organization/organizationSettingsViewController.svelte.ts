import {
	archiveTeamMembershipMutation,
	archiveTeamMutation,
	createTeamMembershipMutation,
	createTeamMutation,
	listTeamMembershipsOptions,
	listTeamsOptions,
	listUsersOptions,
	type ErrorModel,
	type TeamMembership,
	updateTeamMembershipMutation,
	updateTeamsMutation,
} from "$lib/api";
import { useAuthSessionState } from "$lib/auth.svelte";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context, watch } from "runed";
import { SvelteMap } from "svelte/reactivity";

type SectionName = "teams" | "memberships" | "users";
type TeamRole = TeamMembership["attributes"]["role"];

export class OrganizationSettingsViewController {
	session = useAuthSessionState();
	queryClient = useQueryClient();

	selectedTeamId = $state<string>();
	teamSearch = $state<string>();
	userSearch = $state<string>();

	sectionErrors = new SvelteMap<SectionName, ErrorModel>();

	private listTeamsQueryOpts = $derived(listTeamsOptions({
		query: {
			archived: false,
			search: this.teamSearch,
		},
	}));
	teamsQuery = createQuery(() => this.listTeamsQueryOpts);

	private listUsersQueryOpts = $derived(listUsersOptions({
		query: {
			archived: false,
			search: this.userSearch,
		},
	}));
	usersQuery = createQuery(() => this.listUsersQueryOpts);

	private listMembershipsQueryOpts = $derived(listTeamMembershipsOptions({
		query: {
			teamId: this.selectedTeamId,
			limit: 100,
		},
	}))
	membershipsQuery = createQuery(() => ({
		...this.listMembershipsQueryOpts,
		enabled: !!this.selectedTeamId,
	}));

	teams = $derived(this.teamsQuery.data?.data ?? []);
	users = $derived(this.usersQuery.data?.data ?? []);
	memberships = $derived(this.membershipsQuery.data?.data ?? []);

	orgName = $derived(this.session.org?.attributes.name ?? "");
	isOrgAdmin = $derived(this.session.user?.attributes.isOrgAdmin ?? false);
	selectedTeam = $derived(this.teams.find((team) => team.id === this.selectedTeamId));

	availableUsersForSelectedTeam = $derived.by(() => {
		const memberUserIds = new Set(this.memberships.map((m) => m.attributes.userId));
		return this.users.filter((user) => !memberUserIds.has(user.id));
	});

	createTeamMut = createMutation(() => ({
		...createTeamMutation(),
		onSuccess: async () => {
			this.sectionErrors.delete("teams");
			await this.queryClient.invalidateQueries(this.teamsQuery);
		},
		onError: (err) => this.sectionErrors.set("teams", err),
	}));
	updateTeamMut = createMutation(() => ({
		...updateTeamsMutation(),
		onSuccess: async () => {
			this.sectionErrors.delete("teams");
			await this.queryClient.invalidateQueries(this.teamsQuery);
		},
		onError: (err) => this.sectionErrors.set("teams", err),
	}));
	archiveTeamMut = createMutation(() => ({
		...archiveTeamMutation(),
		onSuccess: async () => {
			this.sectionErrors.delete("teams");
			await this.queryClient.invalidateQueries(this.teamsQuery);
			await this.queryClient.invalidateQueries(this.membershipsQuery);
		},
		onError: (err) => this.sectionErrors.set("teams", err),
	}));

	createMembershipMut = createMutation(() => ({
		...createTeamMembershipMutation(),
		onSuccess: async () => {
			this.sectionErrors.delete("memberships");
			this.sectionErrors.delete("users");
			await this.queryClient.invalidateQueries(this.membershipsQuery);
		},
		onError: (err) => this.sectionErrors.set("memberships", err),
	}));
	updateMembershipMut = createMutation(() => ({
		...updateTeamMembershipMutation(),
		onSuccess: async () => {
			this.sectionErrors.delete("memberships");
			this.sectionErrors.delete("users");
			await this.queryClient.invalidateQueries(this.membershipsQuery);
		},
		onError: (err) => this.sectionErrors.set("memberships", err),
	}));
	archiveMembershipMut = createMutation(() => ({
		...archiveTeamMembershipMutation(),
		onSuccess: async () => {
			this.sectionErrors.delete("memberships");
			this.sectionErrors.delete("users");
			await this.queryClient.invalidateQueries(this.membershipsQuery);
		},
		onError: (err) => this.sectionErrors.set("memberships", err),
	}));

	constructor() {
		watch(
			() => this.teams,
			(teams) => {
				if (teams.length === 0) {
					this.selectedTeamId = undefined;
					return;
				}
				if (!this.selectedTeamId || !teams.some((team) => team.id === this.selectedTeamId)) {
					this.selectedTeamId = teams[0].id;
				}
			}
		);
	}

	setTeamSearch(search?: string) {
		this.teamSearch = search || undefined;
	}

	setUserSearch(search?: string) {
		this.userSearch = search || undefined;
	}

	setSelectedTeamId(teamId?: string) {
		this.selectedTeamId = teamId;
	}

	createTeam(name: string) {
		if (!this.isOrgAdmin) return;
		const cleanName = name.trim();
		if (!cleanName) return;
		this.createTeamMut.mutate({
			body: { attributes: { name: cleanName } },
		});
	}

	renameTeam(teamId: string, name: string) {
		if (!this.isOrgAdmin) return;
		const cleanName = name.trim();
		if (!cleanName) return;
		this.updateTeamMut.mutate({
			path: { id: teamId },
			body: { attributes: { name: cleanName } },
		});
	}

	archiveTeam(teamId: string) {
		if (!this.isOrgAdmin) return;
		this.archiveTeamMut.mutate({ path: { id: teamId } });
	}

	addMembership(userId: string, role: TeamRole = "member") {
		if (!this.isOrgAdmin || !this.selectedTeamId) return;
		this.createMembershipMut.mutate({
			body: {
				attributes: {
					teamId: this.selectedTeamId,
					userId,
					role,
				},
			},
		});
	}

	updateMembershipRole(id: string, role: TeamRole) {
		if (!this.isOrgAdmin) return;
		this.updateMembershipMut.mutate({
			path: { id },
			body: {
				attributes: {
					role,
				},
			},
		});
	}

	removeMembership(id: string) {
		if (!this.isOrgAdmin) return;
		this.archiveMembershipMut.mutate({ path: { id } });
	}

	teamMembershipForUser(userId: string): TeamMembership | undefined {
		if (!this.selectedTeamId) return undefined;
		return this.memberships.find(
			m =>
				m.attributes.userId === userId &&
				m.attributes.teamId === this.selectedTeamId
		);
	}

	membershipsForUser(userId: string): TeamMembership[] {
		return this.memberships.filter(m => m.attributes.userId === userId);
	}

	dismissSectionErrorAlert(section: SectionName) {
		this.sectionErrors.delete(section);
	}
}

const ctx = new Context<OrganizationSettingsViewController>("OrganizationSettingsViewController");

export const initOrganizationSettingsViewController = () =>
	ctx.set(new OrganizationSettingsViewController());

export const useOrganizationSettingsViewController = () => ctx.get();
