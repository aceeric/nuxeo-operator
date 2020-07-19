package nuxeo

import "nuxeo-operator/pkg/apis/nuxeo/v1alpha1"

// wip:
// $ ./stream.sh lag -k
//   - verifies this connects nuxeo to strimzi
// todo-me
// 1) support encryption setting - only option is "tls"
// 2) support auth setting: "anonymous" (what's below - default), "sasl-plain", "sasl-scram-sha-512", "tls""
// 3) support user setting - the KafkUser
// E.g.:
//   backingServices:
//   - preConfigured:
//       type: Strimzi
//       resource: strimzi
//       settings:
//         user: nuxeo
//         auth: sasl-scram-sha-512
//         encryption: tls
// ... or:
//       settings:
//         auth: tls # automatically enables tls encryption
// ... or:
//       settings:
//         user: nuxeo  # if user specified without auth, then auth="sasl-plain"
func strimziBacking(preCfg v1alpha1.PreconfiguredBackingService) v1alpha1.BackingService {
	bsvc := v1alpha1.BackingService{
		Name: "strimzi",
		NuxeoConf: "kafka.enabled=true\n" +
			"kafka.bootstrap.servers=" + preCfg.Resource + "-kafka-bootstrap:9092\n" +
			"nuxeo.stream.work.enabled=true\n" +
			"nuxeo.pubsub.provider=stream\n",
	}
	return bsvc
}
