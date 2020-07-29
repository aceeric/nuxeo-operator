package nuxeo

import (
	goerrors "errors"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"nuxeo-operator/pkg/apis/nuxeo/v1alpha1"
)

func crunchyBacking(preCfg v1alpha1.PreconfiguredBackingService) (v1alpha1.BackingService, error) {
	opts, err := parsePreconfigOpts(preCfg)
	if err != nil {
		return v1alpha1.BackingService{}, err
	}
	user, ok := opts["user"]
	if !ok {
		return v1alpha1.BackingService{}, goerrors.New("User is required for Cunchy pre-configured backing service")
	}
	bsvc := v1alpha1.BackingService{
		Name: "crunchy",
		Resources: []v1alpha1.BackingServiceResource{{
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
		}},
		NuxeoConf: "nuxeo.db.host=" + preCfg.Resource + "\n" +
			"nuxeo.db.port=5432\n" +
			"nuxeo.db.name=nuxeo\n" +
			"nuxeo.db.user=${env:PGUSER}\n" +
			"nuxeo.db.password=${env:PGPASSWORD}\n",
		Template: "postgresql",
	}
	return bsvc, nil
}
