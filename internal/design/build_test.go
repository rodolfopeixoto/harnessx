// SPDX-License-Identifier: MIT

package design

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBuild_WritesAllProductMaps(t *testing.T) {
	root := t.TempDir()
	src := filepath.Join(root, "design")
	writeFile(t, src, "index.html",
		`<!doctype html><html><head><title>H</title></head><body>
		<a href="/signup">Sign up</a>
		<form onsubmit="x()"><input class="ui-input"/></form>
		</body></html>`)

	res, err := Build(BuildOptions{Root: root, Source: src})
	require.NoError(t, err)
	require.NotNil(t, res.Manifest)
	require.NotEmpty(t, res.Manifest.Pages)

	for _, p := range []string{
		res.ManifestPath, res.FeatureMapPath, res.ToggleMapPath,
		res.RoadmapPath, res.APIContractsPath, res.FlowMapPath,
	} {
		_, err := os.Stat(p)
		require.NoErrorf(t, err, "missing %s", p)
	}
}

func TestPromoteToggleMap_PreservesStatus(t *testing.T) {
	fm := FeatureMap{
		Features: map[string]FeatureSpec{
			"feature.x": {Status: StatusMock, Routes: []string{"/x"}, BackendRequired: true, APIContract: "POST /api/x"},
		},
	}
	tm := PromoteToggleMap(fm)
	require.Contains(t, tm.Toggles, "feature.x")
	require.Equal(t, StatusMock, tm.Toggles["feature.x"].Status)
	require.Contains(t, tm.Toggles["feature.x"].Description, "needs backend")
}

func TestBuildAPIContracts_OnlyBackendRequired(t *testing.T) {
	fm := FeatureMap{
		Features: map[string]FeatureSpec{
			"feature.static": {Status: StatusStatic, BackendRequired: false},
			"feature.signup": {Status: StatusMock, BackendRequired: true, APIContract: "POST /api/signup"},
		},
	}
	a := BuildAPIContracts(fm)
	require.Len(t, a.Endpoints, 1)
	require.Equal(t, "POST", a.Endpoints[0].Method)
	require.Equal(t, "feature.signup", a.Endpoints[0].Feature)
}

func TestBuildFlowMap_FromManifest(t *testing.T) {
	m := &Manifest{DetectedFlows: []string{"home → /signup", "home → /products"}}
	got := BuildFlowMap(m)
	require.Len(t, got.Flows, 2)
}
