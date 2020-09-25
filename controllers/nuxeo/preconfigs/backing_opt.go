package preconfigs

import (
	"fmt"
	"strings"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
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
		"ca":   {}, // a secret containing key 'ca.crt' for one-way tls
		"tls":  {}, // a secret containing keys 'tls.crt' and 'tls.key' for mutual tls
	},
	v1alpha1.MongoEnterprise: {
		// no options presently - support mongo topologies, auth, encryption in a future operator release
	},
}

// ParsePreconfigOpts parses the options in the passed preconfigured backing service
func ParsePreconfigOpts(preconfigured v1alpha1.PreconfiguredBackingService) (map[string]string, error) {
	return preConfigOpts(preconfigured.Type, preconfigured.Settings)
}

// worker called by ParsePreconfigOpts. The crSetting arg holds what the configurer provided in the
// Nuxeo CR.
func preConfigOpts(typ v1alpha1.PreconfigType, crSetting map[string]string) (map[string]string, error) {
	toReturn := map[string]string{}
	known, ok := opts[typ]
	if !ok {
		return nil, fmt.Errorf("unknown pre-config type: '%v'", typ)
	}
OUTER:
	for k, v := range crSetting {
		cfg := strings.ToLower(k)
		thisSetting, ok := known[cfg]
		if !ok {
			return nil, fmt.Errorf("unknown setting: '%v'", cfg)
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
			return nil, fmt.Errorf("unsupported setting value '%v' for '%v'", val, cfg)
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
			return nil, fmt.Errorf("user not allowed for anonymous Strimzi auth")
		} else if (auth != "anonymous" && auth != "") && user == "" {
			return nil, fmt.Errorf("user required for Strimzi sasl or tls auth")
		}
	case v1alpha1.ECK:
		fallthrough
	case v1alpha1.MongoEnterprise:
		// no additional validations
		return opts, nil
	case v1alpha1.Crunchy:
		user, _ := opts["user"]
		ca, _ := opts["ca"]
		tls, _ := opts["tls"]
		switch {
		case user != "" && ca == "" && tls == "":
			fallthrough // user/pass auth no encrypt
		case user != "" && ca != "" && tls == "":
			fallthrough // user/pass auth TLS encrypt
		case user == "" && ca != "" && tls != "": // mutual TLS
			return opts, nil
		default:
			return nil, fmt.Errorf("unsupported Crunchy authentication/encryption configuration")
		}
	}
	return opts, nil
}
