package nuxeo

import (
	"context"
	"strconv"
	"testing"

	"github.com/aceeric/nuxeo-operator/api/v1alpha1"
	"github.com/aceeric/nuxeo-operator/controllers/util"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
)

// TestBackingServiceECK defines a Nuxeo CR with one backing service for ECK. It creates two simulated secrets
// like the ECK Operator, and creates a simulated deployment like one that would be generated by the Operator from the
// Nuxeo CR. Then it reconciles the backing services and verifies that the configuration was correct. The reconciliation
// should produce a secondary secret, two environment variables, a volume for the secondary secret, a volume mount
// for the secondary secret keys, and a config map containing the backing service nuxeo.conf settings.
func (suite *backingServiceSuite) TestBackingServiceECK() {
	nux := suite.backingServiceSuiteNewNuxeoES()
	dep := genTestDeploymentForBackingSvc()
	err := createECKSecrets(suite)
	require.Nil(suite.T(), err, "Error creating orphaned PVC")
	nuxeoConf, err := suite.r.configureBackingServices(nux, &dep)
	require.Nil(suite.T(), err, "configureBackingServices returned non-nil")
	require.Equal(suite.T(), suite.nuxeoConf, nuxeoConf, "backing service nuxeo.conf should have been returned")
	secSecretName := nux.Name + "-secondary-" + suite.backing
	secret := corev1.Secret{}
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: secSecretName, Namespace: suite.namespace}, &secret)
	require.Nil(suite.T(), err, "configureBackingServices failed to generate secondary secret")

	// Data:
	//  elastic.ca.jks -> len:1984, cap:1986
	//  elastic.truststore.pass -> len:17, cap:18
	require.Equal(suite.T(), 2, len(secret.Data), "Secondary secret not correctly defined")

	// volumes:
	// - name: backing-elastic
	//   projected:
	//     sources:
	//     - secret:
	//         items:
	//         - key: elastic.ca.jks
	//           path: elastic.ca.jks
	//         - key: elastic.truststore.pass
	//           path: elastic.truststore.pass
	//         name: testnux-secondary-elastic
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes), "Volumes incorrectly defined")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes[0].Projected.Sources), "Volume projections incorrectly defined")

	// volumeMounts:
	// - mountPath: /etc/nuxeo-operator/binding/elastic
	//   name: backing-elastic
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts), "Volume Mounts incorrectly defined")

	// ELASTIC_TS_PASS and ELASTIC_PASSWORD
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Containers[0].Env), "Environment variables incorrectly defined")
}

// Similar to TestBackingServiceECK except uses reconcileNodeSets, which creates an actual deployment, and a
// nuxeo.conf ConfigMap. The test verifies that the nuxeo.conf ConfigMap is poperly generated
func (suite *backingServiceSuite) TestReconcileNodeSetsWithBackingSvc() {
	nux := suite.backingServiceSuiteNewNuxeoES()
	_ = createECKSecrets(suite)
	requeue, err := suite.r.reconcileNodeSets(nux)
	require.Nil(suite.T(), err, "reconcileNodeSets returned non-nil")
	require.Equal(suite.T(), true, requeue, "reconcileNodeSets returned unexpected result")
	dep := appsv1.Deployment{}
	depName := deploymentName(nux, nux.Spec.NodeSets[0])
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: depName, Namespace: suite.namespace}, &dep)
	require.Nil(suite.T(), err, "addOrUpdate did not create secret")
	// one for backing services and one for /docker-entrypoint-initnuxeo.d/nuxeo.conf
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts),
		"should be one vol mount for backing secrets and one for the nuxeo.conf ConfigMap")
	// one for backing service secrets and one for nuxeo.conf ConfigMap
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Volumes),
		"should be one volume for backing secrets and one for the nuxeo.conf ConfigMap")
	nuxeoConf := corev1.ConfigMap{}
	cmName := nuxeoConfCMName(nux, suite.nodeSetName)
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: suite.namespace}, &nuxeoConf)
	require.Nil(suite.T(), err, "nuxeo.conf ConfigMap not created")
	require.Equal(suite.T(), suite.nuxeoConf, nuxeoConf.Data["nuxeo.conf"], "nuxeo.conf data incorrect")
}

// Performs a very basic test of the JSONPath parsing that is used to project values from cluster resources other
// than Secrets and ConfigMaps into the Nuxeo container. Parses a Deployment struct and a Nuxeo struct and
// verifies that that JSONPath parsing works correctly.
func (suite *backingServiceSuite) TestJSONPathParse() {
	var replicas, resName []byte
	var err error
	var replicasInt int
	dep := genTestDeploymentForBackingSvc()
	replicas, err = util.GetJsonPathValue(&dep, "{.spec.replicas}")
	require.Nil(suite.T(), err, "Error parsing replicas")
	replicasInt, err = strconv.Atoi(string(replicas))
	require.Nil(suite.T(), err, "Invalid replicas value")
	require.Equal(suite.T(), *dep.Spec.Replicas, int32(replicasInt), "Invalid replicas value")
	nux := suite.backingServiceSuiteNewNuxeoES()
	resName, err = util.GetJsonPathValue(nux, "{.spec.backingServices[0].resources[0].name}")
	require.Nil(suite.T(), err, "Error parsing nuxeo")
	require.Equal(suite.T(), nux.Spec.BackingServices[0].Resources[0].Name, string(resName), "Nuxeo not parsed")
}

// Same as TestJSONPathParse except obtains the value from a cluster resource based on GVK+Name
func (suite *backingServiceSuite) TestValueFromResource() {
	var err error
	var val []byte
	nux := suite.backingServiceSuiteNewNuxeoES()
	err = suite.r.Create(context.TODO(), nux)
	require.Nil(suite.T(), err, "Error creating Nuxeo CR")
	bsr := v1alpha1.BackingServiceResource{
		GroupVersionKind: metav1.GroupVersionKind{
			Group:   "nuxeo.com",
			Version: "v1alpha1",
			Kind:    "Nuxeo",
		},
		Projections: []v1alpha1.ResourceProjection{{
			From:      "{.spec.backingServices[0].resources[0].name}",
			Transform: v1alpha1.CertTransform{},
		}},
		Name: suite.nuxeoName,
	}
	val, _, err = suite.r.getValueFromResource(bsr, suite.namespace, bsr.Projections[0].From)
	require.Nil(suite.T(), err, "JSONPath parse error")
	require.Equal(suite.T(), nux.Spec.BackingServices[0].Resources[0].Name, string(val), "Nuxeo not parsed")
}

// Tests the function that converts a JSONPath to a Secret Key
func (suite *backingServiceSuite) TestPathToKey() {
	val, err := pathToKey("{.spec.backingServices[0].resources[0].name}")
	require.Nil(suite.T(), err, "pathToKey failed")
	require.Equal(suite.T(), ".spec.backingServices0.resources0.name", val, "pathToKey did not convert correctly")
}

// Creates a Nuxeo CR with backing services from three cluster resources, one of which is a service. Each resource
// specifies a mount. Verifies that the mounts were created correctly, and that a secondary secret was created for the
// service with the correct value.
func (suite *backingServiceSuite) TestProjectionMount() {
	objs := make([]runtime.Object, 3)
	objs[0] = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.passSecret,
			Namespace: suite.namespace,
		},
		StringData: map[string]string{"elastic": suite.password},
		Type:       corev1.SecretTypeOpaque,
	}
	objs[1] = &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.cmName,
			Namespace: suite.namespace,
		},
		Data: map[string]string{"config.setting": suite.configVal},
	}
	objs[2] = &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.serviceName,
			Namespace: suite.namespace,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Name:     "web",
				Protocol: corev1.ProtocolTCP,
				Port:     1,
				TargetPort: intstr.IntOrString{
					Type:   intstr.Int,
					IntVal: suite.targetPort,
				},
			}},
		},
	}
	for _, obj := range objs {
		// create cm
		err := suite.r.Create(context.TODO(), obj)
		require.Nil(suite.T(), err, "couldn't create obj: "+obj.GetObjectKind().GroupVersionKind().Kind)
	}
	nux := suite.backingServiceSuiteNewNuxeoMounts()
	dep := genTestDeploymentForBackingSvc()
	_, err := suite.r.configureBackingServices(nux, &dep)
	require.Nil(suite.T(), err, "configureBackingServices returned non-nil")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes), "Incorrect volume count")
	require.Equal(suite.T(), 3, len(dep.Spec.Template.Spec.Volumes[0].Projected.Sources), "Incorrect projection count")
	secondary := &corev1.Secret{}
	secSecretName := nux.Name + "-secondary-" + suite.backing
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: secSecretName, Namespace: suite.namespace}, secondary)
	require.Nil(suite.T(), err, "Secondary secret not created")
	val, ok := secondary.Data[".spec.ports0.targetPort"] // key from jsonPath
	require.True(suite.T(), ok, "Secondary secret not created correctly")
	vali, err := strconv.Atoi(string(val))
	require.Nil(suite.T(), err, "Invalid port value")
	require.Equal(suite.T(), suite.targetPort, int32(vali), "Secondary secret not created with correct value")
}

// Tests addVolumeProjectionAndItems to ensure it handles adds and merges
func (suite *backingServiceSuite) TestProjectionMount2() {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: NuxeoServiceAccountName,
					Volumes: []corev1.Volume{{
						Name: "foo",
						VolumeSource: corev1.VolumeSource{
							Projected: &corev1.ProjectedVolumeSource{
								Sources: []corev1.VolumeProjection{{
									Secret: &corev1.SecretProjection{
										LocalObjectReference: corev1.LocalObjectReference{Name: "foo"},
										Items: []corev1.KeyToPath{{
											Key:  "foo",
											Path: "bar",
										}},
									},
								}},
							},
						},
					}},
				},
			},
		},
	}
	shouldMergeKeys := corev1.Volume{
		Name: "foo",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{{
					Secret: &corev1.SecretProjection{
						LocalObjectReference: corev1.LocalObjectReference{Name: "foo"},
						Items: []corev1.KeyToPath{{
							Key:  "ting",
							Path: "tang",
						}},
					},
				}},
			},
		},
	}
	shouldAddProjection := corev1.Volume{
		Name: "foo",
		VolumeSource: corev1.VolumeSource{
			Projected: &corev1.ProjectedVolumeSource{
				Sources: []corev1.VolumeProjection{{
					ConfigMap: &corev1.ConfigMapProjection{
						LocalObjectReference: corev1.LocalObjectReference{Name: "foo"},
						Items: []corev1.KeyToPath{{
							Key:  "ting",
							Path: "tang",
						}},
					},
				}},
			},
		},
	}
	err := addVolumeProjectionAndItems(&dep, shouldMergeKeys)
	require.Nil(suite.T(), err, "addVolumeProjectionAndItems returned non-nil")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes), "Volume should not be added")
	err = addVolumeProjectionAndItems(&dep, shouldAddProjection)
	require.Nil(suite.T(), err, "addVolumeProjectionAndItems returned non-nil")
	require.Equal(suite.T(), 1, len(dep.Spec.Template.Spec.Volumes), "Volume should not be added")
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Volumes[0].Projected.Sources), "Projection was not added")
}

// Same as TestReconcileNodeSetsWithBackingSvc excepts using a pre-configured backing service rather than
// explicit (verbose) resources and projections.
func (suite *backingServiceSuite) TestPreConfig() {
	nux := &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.nodeSetName,
				Replicas: 1,
			}},
			BackingServices: []v1alpha1.BackingService{{
				Preconfigured: v1alpha1.PreconfiguredBackingService{
					Type:     "ECK",
					Resource: "elastic",
				},
			}},
		},
	}
	_ = createECKSecrets(suite)
	requeue, err := suite.r.reconcileNodeSets(nux)
	require.Nil(suite.T(), err, "reconcileNodeSets returned non-nil")
	require.Equal(suite.T(), true, requeue, "reconcileNodeSets returned unexpected result")
	dep := appsv1.Deployment{}
	depName := deploymentName(nux, nux.Spec.NodeSets[0])
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: depName, Namespace: suite.namespace}, &dep)
	require.Nil(suite.T(), err, "addOrUpdate did not create secret")
	// one for backing services and one for /docker-entrypoint-initnuxeo.d/nuxeo.conf
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Containers[0].VolumeMounts),
		"should be one vol mount for backing secrets and one for the nuxeo.conf ConfigMap")
	// one for backing service secrets and one for nuxeo.conf ConfigMap
	require.Equal(suite.T(), 2, len(dep.Spec.Template.Spec.Volumes),
		"should be one volume for backing secrets and one for the nuxeo.conf ConfigMap")
	nuxeoConf := corev1.ConfigMap{}
	cmName := nuxeoConfCMName(nux, suite.nodeSetName)
	err = suite.r.Get(context.TODO(), types.NamespacedName{Name: cmName, Namespace: suite.namespace}, &nuxeoConf)
	require.Nil(suite.T(), err, "nuxeo.conf ConfigMap not created")
	require.Equal(suite.T(), suite.nuxeoConf, nuxeoConf.Data["nuxeo.conf"], "nuxeo.conf data incorrect")
}

// TestEnvVal tests the ability to get a value from an upstream resource and project it into the Nuxeo container
// as an environment variable with a direct value rather than as a ValueFrom. E.g.:
//  containers:
//  - name: nuxeo
//    env:
//    - name: DATABASE_PORT
//      value: 5432
// This is useful for upstream resource values that aren't sensitive. This test case creates a Nuxeo CR with a
// backing service resource projection that references itself as the upstream resource. This would never happen
// in the real world but it is convenient because it only requires one resource (the Nuxeo CR) to be created
// for the test.
func (suite *backingServiceSuite) TestEnvVal() {
	nux := suite.backingServiceSuiteNewNuxeoEnvVal()
	err := suite.r.Create(context.TODO(), nux)
	require.Nil(suite.T(), err, "Error creating Nuxeo CR")
	dep := genTestDeploymentForBackingSvc()
	_, err = suite.r.configureBackingServices(nux, &dep)
	require.Nil(suite.T(), err, "configureBackingServices returned non-nil")
	expectedEnv := corev1.EnvVar{
		Name:  "BACKING_NAME",
		Value: suite.backing,
	}
	require.Equal(suite.T(), expectedEnv, dep.Spec.Template.Spec.Containers[0].Env[0])
}

// backingServiceSuite is the BackingService test suite structure
type backingServiceSuite struct {
	suite.Suite
	r           NuxeoReconciler
	nuxeoName   string
	namespace   string
	caSecret    string
	cmName      string
	serviceName string
	passSecret  string
	password    string
	caCert      string
	configVal   string
	targetPort  int32
	backing     string
	nuxeoConf   string
	nodeSetName string
}

// SetupSuite initializes the Fake client, a NuxeoReconciler struct, and various test suite constants
func (suite *backingServiceSuite) SetupSuite() {
	suite.r = initUnitTestReconcile()
	suite.nuxeoName = "testnux"
	suite.namespace = "testns"
	suite.caSecret = "elastic-es-http-certs-public"
	suite.passSecret = "elastic-es-elastic-user"
	suite.cmName = "some-config-map"
	suite.serviceName = "test-service"
	suite.password = "testing123"
	suite.caCert = esCaCert()
	suite.configVal = "abc123"
	suite.targetPort = 987
	suite.backing = "elastic"
	suite.nuxeoConf = "elasticsearch.client=RestClient\n" +
		"elasticsearch.restClient.password=${env:ELASTIC_PASSWORD}\n" +
		"elasticsearch.addressList=https://elastic-es-http:9200\n" +
		"elasticsearch.restClient.truststore.path=" + backingMountBase + "elastic/elastic.ca.jks\n" +
		"elasticsearch.restClient.truststore.password=${env:ELASTIC_TS_PASS}\n" +
		"elasticsearch.restClient.truststore.type=JKS\n" +
		"elasticsearch.restClient.username=elastic\n"
	suite.nodeSetName = "test123"
}

// AfterTest removes objects of the type being tested in this suite after each test
func (suite *backingServiceSuite) AfterTest(_, _ string) {
	obj := v1alpha1.Nuxeo{}
	_ = suite.r.DeleteAllOf(context.TODO(), &obj)
	objSecret := corev1.Secret{}
	_ = suite.r.DeleteAllOf(context.TODO(), &objSecret)
	objDep := appsv1.Deployment{}
	_ = suite.r.DeleteAllOf(context.TODO(), &objDep)
}

// This function runs the BackingService unit test suite. It is called by 'go test' and will call every
// function in this file with a backingServiceSuite receiver that begins with "Test..."
func TestBackingServiceUnitTestSuite(t *testing.T) {
	suite.Run(t, new(backingServiceSuite))
}

// backingServiceSuiteNewNuxeoES creates a test Nuxeo struct with one backing service: ECK
func (suite *backingServiceSuite) backingServiceSuiteNewNuxeoES() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			NodeSets: []v1alpha1.NodeSet{{
				Name:     suite.nodeSetName,
				Replicas: 1,
			}},
			BackingServices: []v1alpha1.BackingService{{
				Name: suite.backing,
				Resources: []v1alpha1.BackingServiceResource{{
					GroupVersionKind: metav1.GroupVersionKind{
						Version: "v1",
						Kind:    "Secret",
					},
					Name: suite.caSecret,
					Projections: []v1alpha1.ResourceProjection{{
						Transform: v1alpha1.CertTransform{
							Type:     v1alpha1.TrustStore,
							Cert:     "tls.crt",
							Store:    "elastic.ca.jks",
							Password: "elastic.truststore.pass",
							PassEnv:  "ELASTIC_TS_PASS",
						},
					}},
				}, {
					GroupVersionKind: metav1.GroupVersionKind{
						Version: "v1",
						Kind:    "Secret",
					},
					Name: suite.passSecret,
					Projections: []v1alpha1.ResourceProjection{{
						From: "elastic",
						Env:  "ELASTIC_PASSWORD",
					}},
				}},
				NuxeoConf: suite.nuxeoConf,
			}},
		},
	}
}

// backingServiceSuiteNewNuxeoMounts generates a Nuxeo CR with backing service bindings from a Secret, a ConfigMap,
// and a Service. Should result in one volume with three projection sources, and one volume mount.
func (suite *backingServiceSuite) backingServiceSuiteNewNuxeoMounts() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			BackingServices: []v1alpha1.BackingService{{
				Name: suite.backing,
				Resources: []v1alpha1.BackingServiceResource{{
					GroupVersionKind: metav1.GroupVersionKind{
						Version: "v1",
						Kind:    "Secret",
					},
					Name: suite.passSecret,
					Projections: []v1alpha1.ResourceProjection{{
						From:  "elastic",
						Mount: "elastic.password",
					}},
				}, {
					GroupVersionKind: metav1.GroupVersionKind{
						Version: "v1",
						Kind:    "ConfigMap",
					},
					Name: suite.cmName,
					Projections: []v1alpha1.ResourceProjection{{
						From:  "config.setting",
						Mount: "config.setting",
					}},
				}, {
					GroupVersionKind: metav1.GroupVersionKind{
						Version: "v1",
						Kind:    "Service",
					},
					Name: suite.serviceName,
					Projections: []v1alpha1.ResourceProjection{{
						From:  "{.spec.ports[0].targetPort}",
						Mount: "service.target.port",
					}},
				}},
			}},
		},
	}
}

func (suite *backingServiceSuite) backingServiceSuiteNewNuxeoEnvVal() *v1alpha1.Nuxeo {
	return &v1alpha1.Nuxeo{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.nuxeoName,
			Namespace: suite.namespace,
		},
		Spec: v1alpha1.NuxeoSpec{
			BackingServices: []v1alpha1.BackingService{{
				Name: suite.backing,
				Resources: []v1alpha1.BackingServiceResource{{
					GroupVersionKind: metav1.GroupVersionKind{
						Group:   "nuxeo.com",
						Version: "v1alpha1",
						Kind:    "Nuxeo",
					},
					// self-referencing
					Name: suite.nuxeoName,
					Projections: []v1alpha1.ResourceProjection{{
						From:  "{.spec.backingServices[0].name}", // e.g.: suite.backing
						Env:   "BACKING_NAME",
						Value: true,
					}},
				}},
			}},
		},
	}
}

// genTestDeploymentForBackingSvc creates and returns a Deployment struct minimally configured to support this suite
func genTestDeploymentForBackingSvc() appsv1.Deployment {
	replicas := int32(1)
	dep := appsv1.Deployment{
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: NuxeoServiceAccountName,
					Containers: []corev1.Container{{
						Name: "nuxeo",
					}},
				},
			},
		},
	}
	return dep
}

// Generate connection secrets the way ECK generates them
func createECKSecrets(suite *backingServiceSuite) error {
	userSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.passSecret,
			Namespace: suite.namespace,
		},
		Data: map[string][]byte{"elastic": []byte(suite.password)},
		Type: corev1.SecretTypeOpaque,
	}
	if err := suite.r.Create(context.TODO(), &userSecret); err != nil {
		return err
	}
	caSecret := corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      suite.caSecret,
			Namespace: suite.namespace,
		},
		Data: map[string][]byte{"tls.crt": []byte(suite.caCert)},
		Type: corev1.SecretTypeOpaque,
	}
	return suite.r.Create(context.TODO(), &caSecret)
}

func esCaCert() string {
	return "" +
		"-----BEGIN CERTIFICATE-----\n" +
		"MIIDmDCCAoCgAwIBAgIRAIIhobOevGOuKRgN8oYgUJwwDQYJKoZIhvcNAQELBQAw\n" +
		"KTEQMA4GA1UECxMHZWxhc3RpYzEVMBMGA1UEAxMMZWxhc3RpYy1odHRwMB4XDTIw\n" +
		"MDcxMjIwMzc0M1oXDTIxMDcxMjIwNDc0M1owPTEQMA4GA1UECxMHZWxhc3RpYzEp\n" +
		"MCcGA1UEAxMgZWxhc3RpYy1lcy1odHRwLmJhY2tpbmcuZXMubG9jYWwwggEiMA0G\n" +
		"CSqGSIb3DQEBAQUAA4IBDwAwggEKAoIBAQDBx1arym+A6ek0j5zTmqCfMCTgTPwo\n" +
		"hpY7vfISn780ErjK/gfWPp1aXYEEhvT81OuR8yadPiZJiN6wBQzOz2Ja9VlX/Uy2\n" +
		"4AQqKDWL4VCYHaG8HIsxGFlkqJQfIhKGljhnRri37lBhimoDvUAr/pZgZ2LHeTqm\n" +
		"IkHNXW/7AH9yCH39VQfVVNpfsvD0vjOZDuvKXYf1J5Mz7FYvtbYb8azEfUSF5bE6\n" +
		"lmgaW5KyyeT66zKQKoFeKzr6QVqtImAo9n41TKmm7ztxmCXQQLPoYrAcYWG8qjMI\n" +
		"nsa2ews4sJzSBVWsPi274/Ca67ypER97XxbiQ88VSvLeY21TpE4B5oH/AgMBAAGj\n" +
		"gaYwgaMwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMBBggrBgEF\n" +
		"BQcDAjByBgNVHREEazBpgiBlbGFzdGljLWVzLWh0dHAuYmFja2luZy5lcy5sb2Nh\n" +
		"bIIPZWxhc3RpYy1lcy1odHRwghtlbGFzdGljLWVzLWh0dHAuYmFja2luZy5zdmOC\n" +
		"F2VsYXN0aWMtZXMtaHR0cC5iYWNraW5nMA0GCSqGSIb3DQEBCwUAA4IBAQCSIWTy\n" +
		"m1s01fhgXAPZ6XUpUZwkxsj0Ah7mndedcFvWIjnLnMHd86ZYa8AeqHiWOlS6zbog\n" +
		"SH2iv6VOXgxHn3Dwsb4DFvg4gIp+3x1+4e/60VmT2OBlLeu998ug4XslRjsqZqYc\n" +
		"YUrSi18C/rlYas98xLihWQf7S57tuYua4u+KzK3XFEOxgkgzWEJDC+BQ9pYZcJ/o\n" +
		"vBo4DB2DiVZyJ+b4x6yglVKGXr6zWGlcjeNflsAPx0H3kMdWRfu+LFMvwP/aWEhU\n" +
		"OjAcCtA75EGuWNUK2JSw6H3w5Zg0x0fH6wrtECZlfD7p5KWFYAW1W/NnlQDbngLA\n" +
		"W96Yx0SrW1jDRziV\n" +
		"-----END CERTIFICATE-----\n" +
		"-----BEGIN CERTIFICATE-----\n" +
		"MIIDHjCCAgagAwIBAgIQEEZD9zl4FpfPR6I4MPHFRTANBgkqhkiG9w0BAQsFADAp\n" +
		"MRAwDgYDVQQLEwdlbGFzdGljMRUwEwYDVQQDEwxlbGFzdGljLWh0dHAwHhcNMjAw\n" +
		"NzEyMjAzNzQzWhcNMjEwNzEyMjA0NzQzWjApMRAwDgYDVQQLEwdlbGFzdGljMRUw\n" +
		"EwYDVQQDEwxlbGFzdGljLWh0dHAwggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK\n" +
		"AoIBAQDQz+K8UDC8Qjb7SAXd6X805i/5TiYbWhrKRD87pBXkksdQJ4I7S/0fLpb7\n" +
		"Wn0a2oQ8A6bIWxG8Vt6V3xWgbeQd6u0Vxqvc471Ey9j43CiAZ1kCFzB7nXm2z0fL\n" +
		"kF8HhO1uUSsVt+eRbiw8vxOkjqDKWRADyz71p9ihqaNNb+3CAEAl0n3qK9GjJrFD\n" +
		"dJfktanEzM98kK+ZC/CrAeLmh9w4UBsA07OVgDMDXX4sQAsCTP9HnJAVVVt3bhac\n" +
		"izXq1+sshRhlnBvZB5ulAkzck55QpdFQXCWjJdayUe1dho3H/PeGCbRezSzyx1er\n" +
		"UXcdcHC6ebpAZNo9610nqhm9d4ZdAgMBAAGjQjBAMA4GA1UdDwEB/wQEAwIChDAd\n" +
		"BgNVHSUEFjAUBggrBgEFBQcDAQYIKwYBBQUHAwIwDwYDVR0TAQH/BAUwAwEB/zAN\n" +
		"BgkqhkiG9w0BAQsFAAOCAQEAH+dacGY+MLbAi3eb4a6SUKKJxD+5GmBBfNGbFPrP\n" +
		"j+2mJF7Gj5t/AjRrNbtzDMijdyAxaAE3sTZE6OwSEj6t+K9pwn1RutUgEBpcXU3v\n" +
		"0qL4ZBNeJejlxEKOme+aW5JWSQ9FBaemxntZhe9UebvphD6cxFQNl9fYsInnORnD\n" +
		"6FaD8s6Qd16viWrrj+blrg6jYozsCTzi9wDEwFLwsR1rkJYDIJA8g65v5I5BryMu\n" +
		"G7yx1ZUbM5FW350vtczOnLtD/xm4n1jY9M5xTVFDskJO1IBZLxdLjrSoUJc/6upb\n" +
		"B7kaNdr6ckmmy1HDE3ezg4ca9ufxm6QuBvesPfGUG5Ycqg==\n" +
		"-----END CERTIFICATE-----"
}