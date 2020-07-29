package nuxeo

import (
	goerrors "errors"
	"fmt"
	"strings"

	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
)

// Defines valid settings and values for each pre-configured backing service. E.g. Strimzi encryption has only
// one valid value: tls. If a setting has no values in its array, then there is no validation on that setting.
// E.g.: "user: nuxeo" is valid but "auth: nuxeo" is not.
var opts = map[v1alpha1.PreconfigType]map[string][]string{
	v1alpha1.ECK: {
		"user": {}, // a secret containing keys 'user' and 'password'
	},
	v1alpha1.Strimzi: {
		"auth": {"anonymous", "scram-sha-512", "tls"},
		"user": {}, // a secret whose name is the username containing key 'password'
	},
	v1alpha1.Crunchy: {
		"user": {}, // a secret containing keys 'username' and 'password'
	},
}

// parsePreconfigOpts parses the options in the passed preconfigured backing service
func parsePreconfigOpts(preconfigured v1alpha1.PreconfiguredBackingService) (map[string]string, error) {
	return preConfigOpts(preconfigured.Type, preconfigured.Settings)
}

// worker called by parsePreconfigOpts. The crSetting arg holds what the configurer provided in the
// Nuxeo CR.
func preConfigOpts(typ v1alpha1.PreconfigType, crSetting map[string]string) (map[string]string, error) {
	toReturn := map[string]string{}
	known, ok := opts[typ]
	if !ok {
		return nil, goerrors.New(fmt.Sprintf("unknown pre-config type: '%v'", typ))
	}
OUTER:
	for k, v := range crSetting {
		cfg := strings.ToLower(k)
		thisSetting, ok := known[cfg]
		if !ok {
			return nil, goerrors.New(fmt.Sprintf("unknown setting: '%v'", cfg))
		}
		if len(thisSetting) == 0 { // no validation - all values ok
			toReturn[cfg] = v
		} else {
			val := strings.ToLower(v)
			for _, validSetting := range thisSetting {
				if val == validSetting {
					toReturn[cfg] = string(val)
					continue OUTER
				}
			}
			return nil, goerrors.New(fmt.Sprintf("unsupported setting value '%v' for '%v'", val, cfg))
		}
	}
	return validatePreConfig(typ, toReturn)
}

// Handles cross-validation between configuration options
func validatePreConfig(typ v1alpha1.PreconfigType, opts map[string]string) (map[string]string, error) {
	switch typ {
	case v1alpha1.Strimzi:
		auth, _ := opts["auth"]
		user, _ := opts["user"]
		if (auth == "anonymous" || auth == "") && user != "" {
			return nil, goerrors.New("user not allowed for anonymous Strimzi auth")
		} else if (auth != "anonymous" && auth != "") && user == "" {
			return nil, goerrors.New("user required for Strimzi sasl or tls auth")
		}
	case v1alpha1.ECK:
		// no additional validations
		return opts, nil
	case v1alpha1.Crunchy:
		// no additional validations
		return opts, nil
	}
	return opts, nil
}
