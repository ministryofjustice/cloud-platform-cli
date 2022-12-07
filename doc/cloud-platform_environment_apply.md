## cloud-platform environment apply

Perform a terraform apply and kubectl apply for a given namespace

### Synopsis


	Perform a kubectl apply and a terraform apply for a given namespace using either -namespace flag or the
	the namespace in the given PR Id/Number

	Along with the mandatory input flag, the below environments variables needs to be set
	TF_VAR_cluster_name - e.g. "cp-1902-02" to get the vpc details for some modules like rds, es
	TF_VAR_cluster_state_bucket - State where the cluster state is stored
	TF_VAR_cluster_state_key - folder name/state key inside the state bucket where cluster state is stored
	TF_VAR_github_owner - Github owner: ministryofjustice
	TF_VAR_github_token - Personal access token with repo scope to push github action secrets
	TF_VAR_kubernetes_cluster - Full name of the Cluster e.g. XXXXXX.gr7.eu-west2.eks.amazonaws.com
	PINGDOM_API_TOKEN - API Token to access pingdom
	PIPELINE_TERRAFORM_STATE_LOCK_TABLE - DynamoDB table where the state lock is stored
	PIPELINE_STATE_BUCKET - State bucket where the environments state is stored e.g cloud-platform-terraform-state
	PIPELINE_STATE_KEY_PREFIX - State key/ folder where the environments terraform state is stored e.g cloud-platform-environments
	PIPELINE_STATE_REGION - State region of the bucket e.g. eu-west-1
	PIPELINE_CLUSTER - Cluster name/folder inside namespaces/ in cloud-platform-environments
	PIPELINE_CLUSTER_STATE - Cluster name/folder inside the state bucket where the environments terraform state is stored
	

```
cloud-platform environment apply [flags]
```

### Examples

```
$ cloud-platform environment apply -n <namespace>

```

### Options

```
      --all-namespaces            Apply all namespaces with -all-namespaces
      --cluster string            folder name under namespaces/ inside cloud-platform-environments repo refering to full cluster name
      --commit-timestamp string   Timestamp of current commit from the environment repo
      --github-token string       Personal access Token from Github 
  -h, --help                      help for apply
      --kubecfg string            path to kubeconfig file (default "/home/runner/.kube/config")
  -n, --namespace string          Namespace which you want to perform the apply
```

### Options inherited from parent commands

```
      --skip-version-check   don't check for updates
```

### SEE ALSO

* [cloud-platform environment](cloud-platform_environment.md)	 - Cloud Platform Environment actions

