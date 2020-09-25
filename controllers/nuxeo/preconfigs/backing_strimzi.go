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

package preconfigs

import (
	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Returns a backing service struct configured to connect to Strimzi Kafka. The resource name in the
// passed pre-config is the name of a 'kafka.strimzi.io' resource in the namespace.
func StrimziBacking(preCfg v1alpha1.PreconfiguredBackingService,
	backingMountBase string) (v1alpha1.BackingService, error) {
	opts, err := ParsePreconfigOpts(preCfg)
	if err != nil {
		return v1alpha1.BackingService{}, err
	}
	bsvc := v1alpha1.BackingService{
		Name:      "strimzi",
		Resources: []v1alpha1.BackingServiceResource{},
		NuxeoConf: "kafka.enabled=true\n",
	}
	auth, _ := opts["auth"]
	user, _ := opts["user"]

	switch {
	case auth == "scram-sha-512":
		bsvc.NuxeoConf += "kafka.ssl=true\n" +
			"kafka.sasl.enabled=true\n" +
			"kafka.truststore.type=PKCS12\n" +
			"kafka.truststore.path=" + backingMountBase + "strimzi/truststore.p12\n" +
			"kafka.truststore.password=${env:KAFKA_TRUSTSTORE_PASS}\n" +
			"kafka.security.protocol=SASL_SSL\n" +
			"kafka.sasl.mechanism=SCRAM-SHA-512\n" +
			"kafka.bootstrap.servers=" + preCfg.Resource + "-kafka-bootstrap:9093\n" +
			"kafka.sasl.jaas.config=org.apache.kafka.common.security.scram.ScramLoginModule required username=\"" +
			user + "\" password=\"${env:KAFKA_USER_PASS}\";\n"

		res := []v1alpha1.BackingServiceResource{{
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "secret",
			},
			// KafkaUser CR = user name = user secret name
			Name: user,
			Projections: []v1alpha1.ResourceProjection{{
				From: "password",
				Env:  "KAFKA_USER_PASS",
			}},
		}, {
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "secret",
			},
			Name: preCfg.Resource + "-cluster-ca-cert",
			Projections: []v1alpha1.ResourceProjection{{
				From: "ca.password",
				Env:  "KAFKA_TRUSTSTORE_PASS",
			}, {
				From:  "ca.p12",
				Mount: "truststore.p12",
			}},
		}}
		bsvc.Resources = append(bsvc.Resources, res...)
		break
	case auth == "tls":
		// Nuxeo will add security.protocol=SSL
		bsvc.NuxeoConf += "kafka.ssl=true\n" +
			"kafka.bootstrap.servers=" + preCfg.Resource + "-kafka-bootstrap:9093\n" +
			"kafka.truststore.type=PKCS12\n" +
			"kafka.truststore.path=" + backingMountBase + "strimzi/truststore.p12\n" +
			"kafka.truststore.password=${env:KAFKA_TRUSTSTORE_PASS}\n" +
			"kafka.keystore.type=PKCS12\n" +
			"kafka.keystore.path=" + backingMountBase + "strimzi/keystore.p12\n" +
			"kafka.keystore.password=${env:KAFKA_KEYSTORE_PASS}\n"

		res := []v1alpha1.BackingServiceResource{{
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "secret",
			},
			Name: preCfg.Resource + "-cluster-ca-cert",
			Projections: []v1alpha1.ResourceProjection{{
				From: "ca.password",
				Env:  "KAFKA_TRUSTSTORE_PASS",
			}, {
				From:  "ca.p12",
				Mount: "truststore.p12",
			}},
		}, {
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "secret",
			},
			// KafkaUser CR = user name = user secret name
			Name: user,
			Projections: []v1alpha1.ResourceProjection{{
				From: "user.password",
				Env:  "KAFKA_KEYSTORE_PASS",
			}, {
				From:  "user.p12",
				Mount: "keystore.p12",
			}},
		}}
		bsvc.Resources = append(bsvc.Resources, res...)
	default:
		bsvc.NuxeoConf += "kafka.bootstrap.servers=" + preCfg.Resource + "-kafka-bootstrap:9092\n"
	}
	return bsvc, nil
}
