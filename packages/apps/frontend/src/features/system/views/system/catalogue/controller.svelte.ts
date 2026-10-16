import { page } from "$app/state";
import { createMutation, createQuery, useQueryClient } from "@tanstack/svelte-query";
import { Context } from "runed";
import { useSearchParams } from "runed/kit";
import { onMount } from "svelte";
import { toast } from "svelte-sonner";
import { z } from "zod";
import {
	confirmKnowledgeProposalMutation,
	getKnowledgeProposalOptions,
	getKnowledgeProposalQueryKey,
	listKnowledgeGraphEntitiesOptions,
	listKnowledgeGraphEntitiesQueryKey,
	listKnowledgeProposalsOptions,
	listKnowledgeProposalsQueryKey,
	rejectKnowledgeProposalMutation,
	type ApiError,
	type KnowledgeProposal,
} from "$lib/api";
import { createPaginatedQuery } from "$lib/api/queryPaginator.svelte";
import {
	confirmQuestion,
	entityRowView,
	impliedProposalViews,
	possibleMatchesByEntity,
	proposalView,
} from "$features/system/lib/catalogue/proposals";

export const catalogueTabs = ["services", "repositories"] as const;
export type CatalogueTab = (typeof catalogueTabs)[number];

const catalogueParamsSchema = z.object({
	tab: z.enum(catalogueTabs).default("services"),
	q: z.string().default(""),
});

const isCatalogueTab = (value: string): value is CatalogueTab =>
	(catalogueTabs as readonly string[]).includes(value);

const tabEntityFilters: Record<CatalogueTab, { category: ["container" | "code"]; kind: [string] }> = {
	services: { category: ["container"], kind: ["service"] },
	repositories: { category: ["code"], kind: ["repository"] },
};

/** Proposals shown in each list on the page; the rest are counted. */
const PROPOSAL_PREVIEW_SIZE = 10;
/** One request covers the proposals naming a page of entities. */
const POSSIBLE_MATCHES_PAGE_SIZE = 50;
const URL_WRITE_DEBOUNCE_MS = 300;

export class CatalogueController {
	private queryClient = useQueryClient();

	constructor() {
		onMount(() => {
			return () => this.catalogueParams.cleanup();
		});
	}

	// The tab and name filter are written to the URL once input pauses. Queries and the page reset read the
	// committed URL, so the entity paginator's own URL write never races a pending one.
	private catalogueParams = useSearchParams(catalogueParamsSchema, {
		debounce: URL_WRITE_DEBOUNCE_MS,
		noScroll: true,
	});

	private committed = $derived.by(() => {
		const parsed = catalogueParamsSchema.safeParse(Object.fromEntries(page.url.searchParams));
		if (!parsed.success) {
			return catalogueParamsSchema.parse({});
		}
		return parsed.data;
	});

	private committedSearch = $derived(this.committed.q.trim());
	private committedKey = $derived(`${this.committed.tab}:${this.committedSearch}`);

	tab = $derived(this.catalogueParams.tab);

	setTab = (tab: string) => {
		if (isCatalogueTab(tab)) {
			this.catalogueParams.tab = tab;
		}
	};

	// The proposal a link points at, such as the situation brief's unconfirmed match, in any status.
	focusedProposalId = $derived(page.url.searchParams.get("proposal") ?? undefined);

	/** Shows a proposal at the top of the catalogue, keeping the tab, filter and page. */
	proposalHref = (proposalId: string): string => {
		const params = new URLSearchParams(page.url.searchParams);
		params.set("proposal", proposalId);
		return `${page.url.pathname}?${params.toString()}`;
	};

	focusedProposalQuery = createQuery(() => ({
		...getKnowledgeProposalOptions({ path: { id: this.focusedProposalId ?? "" } }),
		enabled: !!this.focusedProposalId,
	}));

	focusedProposal = $derived.by(() => {
		const proposal = this.focusedProposalQuery.data?.data;
		if (!proposal) {
			return undefined;
		}
		return proposalView(proposal);
	});

	pendingProposals = createPaginatedQuery({
		source: "local",
		queryOptions: () =>
			listKnowledgeProposalsOptions({
				query: { status: ["pending"], page: 1, pageSize: PROPOSAL_PREVIEW_SIZE },
			}),
	});

	pendingQuery = $derived(this.pendingProposals.query);
	pendingViews = $derived((this.pendingQuery.data?.data ?? []).map((proposal) => proposalView(proposal)));
	pendingTotal = $derived(this.pendingQuery.data?.pagination.total ?? 0);

	showRejected = $state(false);

	rejectedProposals = createPaginatedQuery({
		source: "local",
		queryOptions: () => ({
			...listKnowledgeProposalsOptions({
				query: { status: ["rejected"], page: 1, pageSize: PROPOSAL_PREVIEW_SIZE },
			}),
			enabled: this.showRejected,
		}),
	});

	rejectedQuery = $derived(this.rejectedProposals.query);
	rejectedViews = $derived((this.rejectedQuery.data?.data ?? []).map((proposal) => proposalView(proposal)));
	rejectedTotal = $derived(this.rejectedQuery.data?.pagination.total ?? 0);

	get search(): string {
		return this.catalogueParams.q;
	}

	set search(value: string) {
		this.catalogueParams.q = value;
	}

	entities = createPaginatedQuery({
		queryOptions: (pagination) =>
			listKnowledgeGraphEntitiesOptions({
				query: {
					...tabEntityFilters[this.committed.tab],
					search: this.committedSearch || undefined,
					...pagination,
				},
			}),
		resetWhen: () => this.committedKey,
	});

	entitiesQuery = $derived(this.entities.query);

	emptyEntitiesMessage = $derived.by(() => {
		if (this.committedSearch) {
			return "Nothing matches the filter.";
		}
		if (this.committed.tab === "repositories") {
			return "No repositories yet.";
		}
		return "No services yet.";
	});
	private pageEntityIds = $derived((this.entitiesQuery.data?.data ?? []).map((entity) => entity.id));

	// One request finds the pending proposals naming any entity on the page.
	private possibleMatchesQuery = createQuery(() => ({
		...listKnowledgeProposalsOptions({
			query: {
				status: ["pending"],
				entityId: this.pageEntityIds,
				page: 1,
				pageSize: POSSIBLE_MATCHES_PAGE_SIZE,
			},
		}),
		enabled: this.pageEntityIds.length > 0,
	}));

	private possibleMatches = $derived(possibleMatchesByEntity(this.possibleMatchesQuery.data?.data ?? []));

	entityRows = $derived(
		(this.entitiesQuery.data?.data ?? []).map((entity) => entityRowView(entity, this.possibleMatches))
	);

	// Confirming reads the proposal again, for what the merge would settle, before asking.
	confirmCandidateId = $state<string>();
	confirmError = $state.raw<ApiError>();

	// Always read afresh, so the dialog never asks about a cached state of the proposal.
	private confirmCandidateQuery = createQuery(() => ({
		...getKnowledgeProposalOptions({ path: { id: this.confirmCandidateId ?? "" } }),
		enabled: !!this.confirmCandidateId,
		staleTime: 0,
	}));

	private confirmCandidate = $derived.by(() => {
		const proposal = this.confirmCandidateQuery.data?.data;
		if (proposal?.id !== this.confirmCandidateId) {
			return undefined;
		}
		return proposal;
	});

	private confirmMutation = createMutation(() => confirmKnowledgeProposalMutation());
	confirming = $derived(this.confirmMutation.isPending);

	confirmCandidateLoading = $derived(!!this.confirmCandidateId && this.confirmCandidateQuery.isPending);
	// A refetch after a conflict may show the proposal already confirmed by someone else.
	canMerge = $derived(
		!!this.confirmCandidate &&
			this.confirmCandidate.attributes.status !== "confirmed" &&
			!this.confirmCandidateQuery.isFetching &&
			!this.confirming
	);
	confirmCandidateLoadError = $derived(this.confirmCandidateQuery.error ?? undefined);
	confirmCandidateQuestion = $derived(
		this.confirmCandidate ? confirmQuestion(this.confirmCandidate) : undefined
	);
	confirmCandidateImplies = $derived(impliedProposalViews(this.confirmCandidate?.attributes.implies ?? []));

	requestConfirm = (proposalId: string) => {
		this.confirmCandidateId = proposalId;
		this.confirmError = undefined;
	};

	cancelConfirm = () => {
		if (this.confirming) {
			return;
		}
		this.confirmCandidateId = undefined;
		this.confirmError = undefined;
	};

	confirm = async () => {
		const proposal = this.confirmCandidate;
		if (!proposal || this.confirming) {
			return;
		}

		this.confirmError = undefined;
		try {
			await this.confirmMutation.mutateAsync({ path: { id: proposal.id } });
		} catch (error) {
			this.confirmError = error as ApiError;
			await this.refreshAfterConflict(proposal.id, this.confirmError);
			return;
		}

		this.confirmCandidateId = undefined;
		toast.success(confirmedMessage(proposal));
		await this.refreshAfterDecision(proposal.id);
	};

	private rejectMutation = createMutation(() => rejectKnowledgeProposalMutation());
	rejectingId = $state<string>();
	rejectError = $state.raw<{ proposalId: string; error: ApiError }>();

	reject = async (proposalId: string) => {
		if (this.rejectingId) {
			return;
		}

		this.rejectingId = proposalId;
		this.rejectError = undefined;
		try {
			await this.rejectMutation.mutateAsync({ path: { id: proposalId } });
			toast.success("Rejected the match.");
		} catch (error) {
			this.rejectError = { proposalId, error: error as ApiError };
			this.rejectingId = undefined;
			await this.refreshAfterConflict(proposalId, this.rejectError.error);
			return;
		}

		this.rejectingId = undefined;

		await this.refreshAfterDecision(proposalId);
	};

	rejectErrorFor(proposalId: string): ApiError | undefined {
		if (this.rejectError?.proposalId !== proposalId) {
			return undefined;
		}
		return this.rejectError.error;
	}

	// A conflict means someone else decided first; show what they decided.
	private async refreshAfterConflict(proposalId: string, error: ApiError) {
		if (error.code === "conflict") {
			await this.refreshAfterDecision(proposalId);
		}
	}

	// A decision changes the proposal lists, and a merge changes the entities and their aliases. It can also
	// settle the focused proposal, as an implied confirmation.
	private async refreshAfterDecision(proposalId: string) {
		const proposalIds = new Set([proposalId]);
		if (this.focusedProposalId) {
			proposalIds.add(this.focusedProposalId);
		}
		const proposalRefreshes = [...proposalIds].map((id) =>
			this.queryClient.invalidateQueries({ queryKey: getKnowledgeProposalQueryKey({ path: { id } }) })
		);
		await Promise.all([
			this.queryClient.invalidateQueries({ queryKey: listKnowledgeProposalsQueryKey() }),
			this.queryClient.invalidateQueries({ queryKey: listKnowledgeGraphEntitiesQueryKey() }),
			...proposalRefreshes,
		]);
	}
}

function confirmedMessage(proposal: KnowledgeProposal): string {
	const [first, second] = proposal.attributes.values.map((value) => value.value);
	return `Merged ${first} and ${second}.`;
}

const ctx = new Context<CatalogueController>("CatalogueController");
export const initCatalogueController = () => ctx.set(new CatalogueController());
export const useCatalogueController = () => ctx.get();
