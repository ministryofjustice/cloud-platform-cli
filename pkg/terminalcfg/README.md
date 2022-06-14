# Terminal

`terminal` is a sub-command used to set up a session with all the necessary configuration, needed to make sure both kubernetes and terraform are set to the correct in the selected cluster.

---
## Prerequisites
---
aws configuration: 
  - config and credentials files are required to run this command. they need to be set up in the `.aws` directory under the users home directory.
  - To allow the use of the prompt to work correctly the following needs to be set in your profile. `.bashrc`, `.zshrc` 

```
  if [ -n "$KUBECONFIG" ]; then
    PS1="$KUBE_PS1"
  fi
```

---
## Usage
---
```
Usage:
  cloud-platform terminal [command]

Available Commands:
  live        sets up a terminal for the live environment
  manager     sets up a terminal for the manager environment
  test        sets up a terminal for the chosen test environment

Flags:
  -h, --help   help for terminal
```
---
### cloud-plarform terminal live
---
sets up a terminal that connects to the live cluster, setting necessary environment variables for AWS, KUBERNETES and Terraform

```
$ cloud-platform terminal live
```
---
### cloud-plarform terminal manager
---
sets up a terminal that connects to the manager cluster, setting necessary environment variables for AWS, KUBERNETES and Terraform

```
$ cloud-platform terminal manager
```
---
### cloud-plarform terminal test
---
sets up a terminal that connects to a choosen test cluster, setting necessary environment variables for AWS, KUBERNETES and Terraform

```
$ cloud-platform terminal test
```

When ran a prompted will appear to choose a cluster from the list provided.

```
EKS Test Clusters:
 Cluster:  <cluster name>
 Cluster:  <cluster name>
 Cluster:  <cluster name>
 Cluster:  <cluster name>
Please select a cluster to use:
```
Once selected the script use this cluster name to setup the environment variables.