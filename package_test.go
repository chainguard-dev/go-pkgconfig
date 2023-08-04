// Copyright 2023 Chainguard, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkgconfig

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_LoadPackage(t *testing.T) {
	pkg, err := Load("testdata/libpng.pc")
	require.NoError(t, err)

	require.Equal(t, pkg.Name, "libpng")
	require.Equal(t, pkg.Version, "1.6.40")
	require.Equal(t, pkg.Description, "Loads and saves PNG files")

	require.Equal(t, pkg.RequiresPrivate[0].Identifier, "zlib")
}

func Test_WhitespaceTolerance(t *testing.T) {
	pkg, err := Load("testdata/libpng-whitespace.pc")
	require.NoError(t, err)

	require.Equal(t, pkg.Name, "libpng")
	require.Equal(t, pkg.Version, "1.6.40")
	require.Equal(t, pkg.Description, "Loads and saves PNG files")

	require.Equal(t, pkg.RequiresPrivate[0].Identifier, "zlib")
}
