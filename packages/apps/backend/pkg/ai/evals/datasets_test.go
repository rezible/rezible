package evals

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	rezai "github.com/rezible/rezible/pkg/ai"
)

// judgeSituationCandidateDataset holds reference cases for the judge_situation_candidate workflow, for manual
// inspection with a live model. Datasets and their Genkit index live in testdata/datasets; copy them into
// .genkit/datasets to use them in the Developer UI.
const judgeSituationCandidateDataset = "judge_situation_candidate"

// TestJudgeSituationCandidateDatasetDecodesValidInput checks that every case of the judge dataset decodes
// exactly into the workflow's current input and passes its validation, and that the index describes it.
func TestJudgeSituationCandidateDatasetDecodesValidInput(t *testing.T) {
	datasets := filepath.Join("testdata", "datasets")
	encodedCases, readErr := os.ReadFile(filepath.Join(datasets, judgeSituationCandidateDataset+".json"))
	require.NoError(t, readErr)
	var cases []struct {
		TestCaseID string                     `json:"testCaseId"`
		Input      rezai.SituationJudgeInput  `json:"input"`
		Reference  rezai.SituationJudgeOutput `json:"reference"`
	}
	// Unknown fields fail, so the cases keep the input's current shape.
	decoder := json.NewDecoder(bytes.NewReader(encodedCases))
	decoder.DisallowUnknownFields()
	require.NoError(t, decoder.Decode(&cases))
	require.NotEmpty(t, cases)
	for _, datasetCase := range cases {
		require.NoError(t, datasetCase.Input.Validate(), "case %s", datasetCase.TestCaseID)
	}

	encodedIndex, indexErr := os.ReadFile(filepath.Join(datasets, "index.json"))
	require.NoError(t, indexErr)
	var index map[string]struct {
		DatasetID    string `json:"datasetId"`
		Size         int    `json:"size"`
		TargetAction string `json:"targetAction"`
	}
	require.NoError(t, json.Unmarshal(encodedIndex, &index))
	entry, indexed := index[judgeSituationCandidateDataset]
	require.True(t, indexed, "the dataset is in the index")
	require.Equal(t, judgeSituationCandidateDataset, entry.DatasetID)
	require.Equal(t, len(cases), entry.Size)
	require.Equal(t, "/flow/"+rezai.JudgeSituationCandidateDefinition.Name, entry.TargetAction)
}
