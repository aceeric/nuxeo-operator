package preconfigs

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
)

// Generates a backing service struct for integrating with Crunchy Postgres. Presently some connection
// params are hard-coded.
// todo-me implement projection of values from PgCluster resource into the container
//  as environment variables with static values (e.g. host, service) since these are not sensitive
func CrunchyBacking(preCfg v1alpha1.PreconfiguredBackingService, backingMountBase string) (v1alpha1.BackingService, error) {
	opts, err := ParsePreconfigOpts(preCfg)
	if err != nil {
		return v1alpha1.BackingService{}, err
	}
	user, _ := opts["user"]
	ca, _ := opts["ca"]
	tls, _ := opts["tls"]
	var resources []v1alpha1.BackingServiceResource
	nxconf := "nuxeo.db.host=" + preCfg.Resource + "\n" +
		"nuxeo.db.port=5432\n" +
		"nuxeo.db.name=nuxeo\n"
	bsvc := v1alpha1.BackingService{
		Name:     "crunchy",
		Template: "postgresql",
	}
	if user != "" {
		res := v1alpha1.BackingServiceResource{
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "secret",
			},
			Name: user,
			Projections: []v1alpha1.ResourceProjection{{
				From: "username",
				Env:  "PGUSER",
			}, {
				From: "password",
				Env:  "PGPASSWORD",
			}},
		}
		resources = append(resources, res)
		nxconf += "nuxeo.db.user=${env:PGUSER}\n" +
			"nuxeo.db.password=${env:PGPASSWORD}\n"
	}
	if ca != "" {
		res := v1alpha1.BackingServiceResource{
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "secret",
			},
			Name: ca,
			Projections: []v1alpha1.ResourceProjection{{
				From:  "ca.crt",
				Mount: "ca.crt",
			}},
		}
		resources = append(resources, res)
		if tls == "" {
			// one way TLS
			nxconf += "nuxeo.db.jdbc.url=jdbc:postgresql://${nuxeo.db.host}:5432/nuxeo" +
				"?user=${nuxeo.db.user}&password=${nuxeo.db.password}" +
				"&ssl=true" +
				"&sslmode=verify-ca" +
				"&sslrootcert=" + backingMountBase + "crunchy/ca.crt\n"
		}
	}
	if tls != "" {
		res := v1alpha1.BackingServiceResource{
			GroupVersionKind: metav1.GroupVersionKind{
				Group:   "",
				Version: "v1",
				Kind:    "secret",
			},
			Name: tls,
			Projections: []v1alpha1.ResourceProjection{{
				From:  "tls.crt",
				Mount: "tls.crt",
			}, {
				From:  "tls.key",
				Mount: "tls.key",
			}},
		}
		resources = append(resources, res)
		nxconf += "nuxeo.db.user=\n" +
			"nuxeo.db.password=\n" +
			"nuxeo.db.jdbc.url=jdbc:postgresql://${nuxeo.db.host}:5432/nuxeo" +
			"?ssl=true" +
			"&sslmode=verify-ca" +
			"&sslrootcert=" + backingMountBase + "crunchy/ca.crt" +
			"&sslcert=" + backingMountBase + "crunchy/tls.crt" +
			"&sslkey=" + backingMountBase + "crunchy/tls.key"
	}
	bsvc.Resources = resources
	bsvc.NuxeoConf = nxconf
	return bsvc, nil
}
