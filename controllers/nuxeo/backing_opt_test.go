/*
Copyright 2020 Eric Ace.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package nuxeo

import (
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/nuxeo/preconfigs"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// Tests that valid options for Strimzi can be parsed
func (suite *backingOptSuite) TestBackingOptsGood() {
	goodOpts := map[string]string{
		"auth": "Scram-SHa-512",
		"user": "FOO",
	}
	pbs := v1alpha1.PreconfiguredBackingService{
		Type:     v1alpha1.Strimzi,
		Settings: goodOpts,
	}
	parsed, err := preconfigs.ParsePreconfigOpts(pbs)
	require.Nil(suite.T(), err, "parsePreconfigOpts should not have errored")
	user, ok := parsed["user"]
	require.True(suite.T(), ok, "Did not get a user back")
	require.Equal(suite.T(), "FOO", user, "Did not get a user back correctly")
}

// Tests that an unknown type is rejected. This could happen if a new preconfigured type was added
// and no corresponding settings were added in support
func (suite *backingOptSuite) TestBackingOptsUnknownType() {
	pbs := v1alpha1.PreconfiguredBackingService{
		Type: "Unknown",
	}
	_, err := preconfigs.ParsePreconfigOpts(pbs)
	require.NotNil(suite.T(), err, "parsePreconfigOpts should have errored")
}

// Tests that an invalid value for a known setting is rejected
func (suite *backingOptSuite) TestBackingOptsBad() {
	goodOpts := map[string]string{
		"auth": "this is not valid",
		"user": "FOO",
	}
	pbs := v1alpha1.PreconfiguredBackingService{
		Type:     v1alpha1.Strimzi,
		Settings: goodOpts,
	}
	_, err := preconfigs.ParsePreconfigOpts(pbs)
	require.NotNil(suite.T(), err, "parsePreconfigOpts should have errored")
}

// Tests that an unknown setting is rejected
func (suite *backingOptSuite) TestBackingOptsUnknownSetting() {
	goodOpts := map[string]string{
		"auth":    "anonymous",
		"user":    "FOO",
		"invalid": "",
	}
	pbs := v1alpha1.PreconfiguredBackingService{
		Type:     v1alpha1.Strimzi,
		Settings: goodOpts,
	}
	_, err := preconfigs.ParsePreconfigOpts(pbs)
	require.NotNil(suite.T(), err, "parsePreconfigOpts should not errored")
}

// backingOptSuite is the BackingOpt test suite structure
type backingOptSuite struct {
	suite.Suite
	r NuxeoReconciler
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *backingOptSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *backingOptSuite) AfterTest(_, _ string) {
	// nop
}

// This function runs the BackingOpt unit test suite. It is called by 'go test' and will call every
// function in this file with a backingOptSuite receiver that begins with "Test..."
func TestBackingOptUnitTestSuite(t *testing.T) {
	suite.Run(t, new(backingOptSuite))
}
