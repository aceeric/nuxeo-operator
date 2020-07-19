package nuxeo

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
)

// returns a backing service configured to connect to ECK. The passed resource name is the name of
// an 'elasticsearch.k8s.elastic.co' resource in the namespace.
// todo-me enhance with ability to support anonymous, non-tls, two-way tls using 'settings' in preCfg arg, which
//  is presently ignored
func eckBacking(preCfg v1alpha1.PreconfiguredBackingService) v1alpha1.BackingService {
	const trustStore = "elastic.ca.jks"
	bsvc := v1alpha1.BackingService{
		Name: "elastic",
		// first resource converts the ECK tls.crt into a Java trust store
		Resources: []v1alpha1.BackingServiceResource{{
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "secret",
			},
			Name: preCfg.Resource + "-es-http-certs-public",
			Projections: []v1alpha1.ResourceProjection{{
				Transform: v1alpha1.CertTransform{
					Type:     "TrustStore",
					Cert:     "tls.crt",
					Store:    trustStore,
					Password: "elastic.truststore.pass",
					PassEnv:  "ELASTIC_TS_PASS",
				},
			}},
		}, {
			// second resource projects the 'elastic' user password to env var ELASTIC_PASSWORD
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "secret",
			},
			Name: preCfg.Resource + "-es-elastic-user",
			Projections: []v1alpha1.ResourceProjection{{
				From: "elastic",
				Env: "ELASTIC_PASSWORD",
			}},
		}},
		// nuxeo.conf pulls everything together. The elastic user is always 'elastic'. The URL is always
		// a service named <elasticsearch resource name>-es-http. The trust store and password are generated
		// by the TrustStore transform. Only JKS is presently supported.
		NuxeoConf: "elasticsearch.client=RestClient\n" +
			"elasticsearch.restClient.username=elastic\n" +
			"elasticsearch.restClient.password=${env:ELASTIC_PASSWORD}\n" +
			"elasticsearch.addressList=https://" + preCfg.Resource + "-es-http:9200\n" +
			"elasticsearch.restClient.truststore.path=" + backingMountBase + "elastic/" + trustStore + "\n" +
			"elasticsearch.restClient.truststore.password=${env:ELASTIC_TS_PASS}\n" +
			"elasticsearch.restClient.truststore.type=JKS\n",
	}
	return bsvc
}
