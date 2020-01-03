Terraform Chef Solo Provider v 0.1
==================
 
Terraform ChefSolo Provider is a data-source designed to easily helps you
generating node files. It is a basic extension of the data-source "template_file".

This plugin is supposed to be used with the [Terraform ChefSolo provisioner](https://github.com/criteo-forks/terraform-provisioner-chefsolo)

## Example of usage : 

As an attribute file we use `attributes/master.pl` : 

```json
{
  "ipaddress": "${node_ip}",
  "fqdn": "${node_id}",
  "location": {
    "datacenter": "PAR"
  },
  "network": {
    "setup": {
      "teaming": false
    },
    "restart": false,
    "config": {
      "hosts": ${hosts}
    }
  },
  "vault": {
    "mesos_keytabs": {
      "marathon": {
        "keytab": "SOME8VALUE"
      }
    }
  }
}
```

As a terraform configuration we can now use this : 

```hcl
data "template_chefsolo" "consul_server" {

  # Generate multiple node files with a count
  count = "${length(var.instances_ips)}"

  # Create automatic attributes for your node files and use a TPL file
  automatic_attributes = "${file("${path.module}/attributes/master.tpl")}"
  
  # Specify your node-id, it is required
  node_id         = "${element(var.instances_hostnames, count.index)}"
  
  # Specify policy names and policy groups / run_list
  # it'll be automatically added in a json file
  policy_name     = "consul_server"
  policy_group    = "local"
  environment     = "preprod"


  # Every variables that are declared here can be accessed in TPL files
  # with the syntax `${name_of_my_var}` 
  vars {
    node_ip     = "${element(var.instances_ips, count.index)}"
    node_id     = "${element(var.instances_hostnames, count.index)}"
    hosts       = "${jsonencode(zipmap(
                        var.instances_ips,
                        var.instances_hostnames)
                    )}"
  }
}

# Output Node files , it'll return an array of Rendered json files
output "node_files" {
    value = "${data.template_chefsolo.consul_server.*.node}"
}

# Output DNA files, it'll return an array of Rendered json files
output "dna_files" {
    value = "${data.template_chefsolo.consul_server.*.dna}"
}
```






## Options 

- `default_attributes`:
	- Type:        String
	- Optional:    true
	- Description: "Default attributes of your chef node"

- `automatic_attributes`:
	- Type:        String
	- Optional:    true
	- Description: "Automatic attributes of your chef node"

- `vars`:
	- Type:         Map
	- Optional:     true
	- Default:      {}
	- Description:  "variables to substitute"

- `node_id`:
	- Type:        String
	- Required:    true
	- Description: "Instance ID of the node"

- `policy_name`:
	- Type:          String
	- Optional:      true
	- Description:   "Policy name to use"
	- ConflictsWith: "run_list"

- `policy_group`:
	- Type:          String
	- Optional:      true
	- Default:       "local"
	- Description:   "Policy group to use"
	- ConflictsWith: "run_list"

- `named_run_list`:
	- Type:          String
	- Optional:      true
	- Default:       ""
	- Description:   "Optional named run list to target"
	- ConflictsWith: "run_list"

- `run_list`:
	- Type:          List[String]
	- Optional:      true
	- Description:   "List of cookbooks to run"
	- ConflictsWith: "named_run_list", "policy_group", "policy_name"

- `environment`:
	- Type:        String
	- Optional:    true
	- Default:     "local"
	- Description: "Chef environment"

## Output

- `node`:
	- Type:        String
	- Computed:    true
	- Description: "rendered node"

- `dna`:
	- Type:        String
	- Computed:    true
	- Description: "rendered dna"

- `use_policyfile`:
	- Type:        Bool
	- Computed:    true
	- Description: "rendered bool"
