package outputs_wavefront_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/influxdata/telegraf/config"
	_ "github.com/influxdata/telegraf/migrations/outputs_wavefront" // register migration
	_ "github.com/influxdata/telegraf/plugins/outputs/wavefront"    // register plugin
)

func TestCannotMigrateStringToNumber(t *testing.T) {
	cfg := []byte(`
[[outputs.wavefront]]
  string_to_number = "test_key"
`)
	// Migrate and check that nothing changed
	output, n, err := config.ApplyMigrations(cfg)
	require.NoError(t, err)
	require.Equal(t, cfg, output)
	require.Equal(t, uint64(0), n)
}

func TestCases(t *testing.T) {
	// Get all directories in testdata
	folders, err := os.ReadDir("testcases")
	require.NoError(t, err)

	for _, f := range folders {
		// Only handle folders
		if !f.IsDir() {
			continue
		}

		t.Run(f.Name(), func(t *testing.T) {
			testcasePath := filepath.Join("testcases", f.Name())
			inputFile := filepath.Join(testcasePath, "telegraf.conf")
			expectedFile := filepath.Join(testcasePath, "expected.conf")

			// Read the expected output
			expected := config.NewConfig()
			require.NoError(t, expected.LoadConfig(expectedFile))
			require.NotEmpty(t, expected.Outputs)

			// Read the input data
			input, remote, err := config.LoadConfigFile(inputFile)
			require.NoError(t, err)
			require.False(t, remote)
			require.NotEmpty(t, input)

			// Migrate
			output, n, err := config.ApplyMigrations(input)
			require.NoError(t, err)
			require.NotEmpty(t, output)
			require.GreaterOrEqual(t, n, uint64(1))
			actual := config.NewConfig()
			require.NoError(t, actual.LoadConfigData(output, config.EmptySourcePath))

			// Test the output
			require.Len(t, actual.Outputs, len(expected.Outputs))
			actualIDs := make([]string, 0, len(expected.Outputs))
			expectedIDs := make([]string, 0, len(expected.Outputs))
			for i := range actual.Outputs {
				actualIDs = append(actualIDs, actual.Outputs[i].ID())
				expectedIDs = append(expectedIDs, expected.Outputs[i].ID())
			}
			require.ElementsMatch(t, expectedIDs, actualIDs, string(output))
		})
	}
}
