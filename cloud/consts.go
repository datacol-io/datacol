package cloud

const (
	CloudsqlSecretName    = "cloudsql-secret"
	CloudsqlContainerName = "cloudsql-proxy"
	CloudsqlImage         = "gcr.io/cloudsql-docker/gce-proxy:1.09"
	CloudsqlCredVolName   = "cloudsql-instance-credentials"
	DefaultProcPort       = 8080
)

type CloudProvider string

const (
	AwsProvider   CloudProvider = "aws"
	GCPProvider   CloudProvider = "gcp"
	LocalProvider CloudProvider = "local"
)
