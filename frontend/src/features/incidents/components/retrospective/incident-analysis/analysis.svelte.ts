import { incidentCtx } from "$src/features/incidents/lib/context";
import {
  getSystemAnalysisOptions,
  type SystemAnalysis,
  type SystemAnalysisComponent,
  type SystemAnalysisRelationship,
} from "$src/lib/api";
import { createQuery, useQueryClient } from "@tanstack/svelte-query";
import { watch } from "runed";

const createAnalysisState = () => {
  let analysisId = $state<string>();
  let data = $state<SystemAnalysis>();
  let componentDialogOpen = $state(false);
  let editingComponent = $state<SystemAnalysisComponent>();
  let relationshipDialogOpen = $state(false);
  let editingRelationship = $state<SystemAnalysisRelationship>();

  const setup = () => {
    analysisId = incidentCtx.get().attributes.system_analysis_id;

    const queryClient = useQueryClient();

    const analysisQuery = createQuery(
      () => ({
        ...getSystemAnalysisOptions({ path: { id: analysisId ?? "" } }),
        enabled: !!analysisId,
      }),
      queryClient,
    );

    watch(
      () => analysisQuery.data,
      (body) => {
        if (!body?.data) return;
        data = body.data;
      },
    );
  };

  const setComponentDialogOpen = (
    open: boolean,
    editComponent?: SystemAnalysisComponent,
  ) => {
    componentDialogOpen = open;
    editingComponent = editComponent;
  };

  const setRelationshipDialogOpen = (
    open: boolean,
    editRel?: SystemAnalysisRelationship,
  ) => {
    relationshipDialogOpen = open;
    editingRelationship = editRel;
  };

  return {
    setup,
    get data() {
      return data;
    },
    get componentDialogOpen() {
      return componentDialogOpen;
    },
    setComponentDialogOpen,
    get editingComponent() {
      return editingComponent;
    },
    get relationshipDialogOpen() {
      return relationshipDialogOpen;
    },
    setRelationshipDialogOpen,
    get editingRelationship() {
      return editingRelationship;
    },
  };
};

export const analysis = createAnalysisState();
