package preconfigs

import (
	"fmt"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
)

// This is just a very basic pre-config for mongo.com Enterprise supporting a single stand-alone Mongo
// DB with no client auth. A future version of the operator will support the various Mongo topologies, plus
// TLS encryption and SCRAM/TLS auth. This is just an initial version to get a stake in the sand.
func MongoEntBacking(preCfg v1alpha1.PreconfiguredBackingService, _ string) (v1alpha1.BackingService, error) {
	_, err := ParsePreconfigOpts(preCfg)
	if err != nil {
		return v1alpha1.BackingService{}, err
	}
	mongoUrl := fmt.Sprintf("nuxeo.mongodb.server=mongodb://%[1]v-0.%[1]v-svc.backing.svc.cluster.local:27017\n",
		preCfg.Resource)
	bsvc := v1alpha1.BackingService{
		Name:      "mongoent",
		Template:  "mongodb",
		Resources: []v1alpha1.BackingServiceResource{},
		NuxeoConf: mongoUrl +
			"nuxeo.mongodb.ssl=false\n" +
			"nuxeo.mongodb.dbname=nuxeo",
	}
	return bsvc, nil
}
