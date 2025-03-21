<script lang="ts">
	import { setPageBreadcrumbs } from "$features/app/lib/appShellState.svelte";
	import { page } from "$app/stores";
	import { onMount } from "svelte";
	import type { User } from "$lib/api/oapi.gen/types.gen";
	
	import UserProfile from "./UserProfile.svelte";
	import OncallStats from "./OncallStats.svelte";
	import IncidentParticipation from "./IncidentParticipation.svelte";
	import TeamMembership from "./TeamMembership.svelte";
	import { generateMockOncallShifts, generateMockIncidents, generateMockTeams } from "./utils";
	
	let user: User | null = null;
	let loading = true;
	let error: Error | null = null;
	
	let oncallShifts: any[] = [];
	let incidents: any[] = [];
	let teams: any[] = [];
	
	onMount(async () => {
		try {
			// In a real app, you would fetch the user data from your API
			// const response = await fetch(`/api/users/${$page.params.id}`);
			// if (!response.ok) throw new Error('Failed to fetch user');
			// const data = await response.json();
			// user = data.data;
			
			// For demo purposes, we'll create mock data
			await new Promise(resolve => setTimeout(resolve, 500)); // Simulate network delay
			
			user = {
				id: $page.params.id,
				attributes: {
					name: "Jane Smith",
					email: "jane.smith@example.com",
					chatId: "U123456",
					timezone: "America/New_York",
					createdAt: new Date(Date.now() - 90 * 24 * 60 * 60 * 1000).toISOString() // 90 days ago
				}
			};
			
			// Generate mock data for the user
			oncallShifts = generateMockOncallShifts(user.id);
			incidents = generateMockIncidents(user.id);
			teams = generateMockTeams(user.id);
			
			// Update breadcrumbs with actual user name
			setPageBreadcrumbs(() => [
				{ label: "Users", href: "/users" },
				{ label: user.attributes.name, href: `/users/${user.id}`},
			]);
			
			loading = false;
		} catch (err) {
			console.error("Error fetching user:", err);
			error = err instanceof Error ? err : new Error("Unknown error occurred");
			loading = false;
		}
	});
</script>

<div class="user-page">
	{#if loading}
		<div class="loading-state">
			<div class="spinner"></div>
			<p>Loading user profile...</p>
		</div>
	{:else if error}
		<div class="error-state">
			<h2>Error Loading User</h2>
			<p>{error.message}</p>
			<button class="retry-button" on:click={() => window.location.reload()}>
				Retry
			</button>
		</div>
	{:else if user}
		<UserProfile {user} />
		
		<div class="user-data-grid">
			<OncallStats {user} {oncallShifts} />
			<IncidentParticipation {user} {incidents} />
		</div>
		
		<TeamMembership {user} {teams} />
	{/if}
</div>

<style>
	.user-page {
		max-width: 1200px;
		margin: 0 auto;
		padding: 1.5rem;
	}
	
	.loading-state, .error-state {
		display: flex;
		flex-direction: column;
		align-items: center;
		justify-content: center;
		padding: 4rem 2rem;
		text-align: center;
	}
	
	.spinner {
		width: 40px;
		height: 40px;
		border: 4px solid rgba(0, 0, 0, 0.1);
		border-left-color: var(--color-primary);
		border-radius: 50%;
		animation: spin 1s linear infinite;
		margin-bottom: 1rem;
	}
	
	@keyframes spin {
		to { transform: rotate(360deg); }
	}
	
	.error-state h2 {
		color: var(--color-error);
		margin-bottom: 0.5rem;
	}
	
	.retry-button {
		margin-top: 1rem;
		padding: 0.5rem 1rem;
		background-color: var(--color-primary);
		color: white;
		border: none;
		border-radius: 4px;
		cursor: pointer;
	}
	
	.retry-button:hover {
		background-color: var(--color-primary-dark);
	}
	
	.user-data-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: 1.5rem;
		margin-top: 1.5rem;
	}
	
	@media (min-width: 768px) {
		.user-data-grid {
			grid-template-columns: repeat(2, 1fr);
		}
	}
	
	.card {
		background-color: var(--color-bg-card);
		border-radius: 8px;
		padding: 1.5rem;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
	}
</style>
