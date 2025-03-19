package main

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestReadControlDefaultVersion(t *testing.T) {
	controlF, err := os.CreateTemp(t.TempDir(), "test.control")
	require.NoError(t, err)

	defer controlF.Close() //nolint:errcheck // temporary file for testing

	buf := `comment = 'API surface for DocumentDB for PostgreSQL'
default_version = '0.100-0'
module_pathname = '$libdir/pg_documentdb'
relocatable = false
superuser = true
requires = 'documentdb_core, pg_cron, tsm_system_rows, vector, postgis, rum'`
	_, err = io.WriteString(controlF, buf)
	require.NoError(t, err)

	controlDefaultVersion, err := getControlDefaultVersion(controlF.Name())
	require.NoError(t, err)

	require.Equal(t, "0.100.0", controlDefaultVersion)
}
