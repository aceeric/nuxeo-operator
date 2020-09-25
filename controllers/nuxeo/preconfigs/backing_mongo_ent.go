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
