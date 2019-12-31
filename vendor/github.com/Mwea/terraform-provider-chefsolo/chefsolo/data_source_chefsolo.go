package chefsolo

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"github.com/hashicorp/hil"
	"github.com/hashicorp/terraform/config"
	"github.com/hashicorp/terraform/helper/schema"
	//"regexp"
	"github.com/hashicorp/hil/ast"
	"strings"
)

func dataSourceChefSoloFile() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceChefSoloFileRead,

		Schema: map[string]*schema.Schema{
			"default_attributes": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Contents of the chefsolo",
			},
			"automatic_attributes": {
				Type:        schema.TypeString,
				Optional:    true,
				Description: "Contents of the chefsolo",
			},
			"vars": {
				Type:         schema.TypeMap,
				Optional:     true,
				Default:      make(map[string]interface{}),
				Description:  "variables to substitute",
				ValidateFunc: validateVarsAttribute,
			},
			"node_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Instance ID of the node",
			},
			"policy_name": {
				Type:          schema.TypeString,
				Optional:      true,
				Description:   "Policy name to use",
				ConflictsWith: []string{"run_list"},
			},
			"policy_group": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "local",
				Description:   "Policy group to use",
				ConflictsWith: []string{"run_list"},
			},
			"named_run_list": {
				Type:          schema.TypeString,
				Optional:      true,
				Default:       "",
				Description:   "Optional named run list to target",
				ConflictsWith: []string{"run_list"},
			},
			"run_list": {
				Type:          schema.TypeList,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				Description:   "List of cookbooks to run",
				ConflictsWith: []string{"named_run_list", "policy_group", "policy_name"},
			},
			"environment": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "local",
				Description: "Chef environment",
			},
			"node": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "rendered node",
			},
			"dna": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "rendered dna",
			},
			"use_policyfile": {
				Type:        schema.TypeBool,
				Computed:    true,
				Description: "rendered bool",
			},
		},
	}
}

func dataSourceChefSoloFileRead(d *schema.ResourceData, _ interface{}) error {
	toRender := make(map[string]bool)
	toRender["node"] = true
	toRender["dna"] = false
	for k, v := range toRender {
		rendered, err := renderChefSoloFile(d, v)
		if err != nil {
			return err
		}
		// environment  named_run_list
		d.Set(k, rendered)
		d.SetId(hash(rendered))
	}
	d.Set("use_policyfile", d.Get("policy_name").(string) != "")
	return nil
}

func renderChefSoloFile(d *schema.ResourceData, node bool) (string, error) {
	vars := d.Get("vars").(map[string]interface{})
	vars["id"] = d.Get("node_id")
	fullAttrs, err := createMapAttributes(d, vars, node)
	if err != nil {
		return "", err
	}

	if fullAttrs, err = injectChefSoloVars(d, fullAttrs); err != nil {
		return "", err
	}
	data, err := json.Marshal(fullAttrs)
	if err != nil {
		return "", fmt.Errorf("unable to build json file %v", err)
	}

	rendered, err := executeChefSolo(string(data[:]), vars)
	return rendered, nil
}

func createMapAttributes(d *schema.ResourceData, vars map[string]interface{}, node bool) (map[string]interface{}, error) {
	fullAttrs := make(map[string]interface{})

	types := []string{"automatic", "default"}

	for _, level := range types {
		if attrRaw, ok := d.GetOk(level + "_attributes"); ok {
			rendered, err := executeChefSolo(attrRaw.(string), vars)
			if err != nil {
				return nil, fmt.Errorf("unable to evaluate dynamic attributes %s: %v", attrRaw, err)
			}
			var attrs map[string]interface{}
			if err := json.Unmarshal([]byte(rendered), &attrs); err != nil {
				return nil, fmt.Errorf("error parsing %s_attributes: %v", level, err)
			}
			if node {
				fullAttrs[level] = attrs
			} else {
				for k, v := range attrs {
					fullAttrs[k] = v
				}
			}
		}
	}
	return fullAttrs, nil
}

func injectChefSoloVars(d *schema.ResourceData, vars map[string]interface{}) (map[string]interface{}, error) {
	runList := d.Get("run_list").([]interface{})
	policyGroup := d.Get("policy_group").(string)
	policyName := d.Get("policy_name").(string)

	vars["id"] = d.Get("node_id")
	if (policyGroup == "" || policyName == "") && len(runList) <= 0 {
		return nil, fmt.Errorf("neither run_list or policy_name/policy_group has been set")
	}

	if policyName != "" {
		vars["policy_name"] = policyName
		vars["policy_group"] = policyGroup
	} else {
		vars["run_list"] = runList
	}
	return vars, nil
}

// execute parses and executes a chefsolo using vars.
func executeChefSolo(s string, vars map[string]interface{}) (string, error) {
	root, err := hil.Parse(s)
	if err != nil {
		return "", err
	}

	varmap := make(map[string]ast.Variable)
	for k, v := range vars {
		// As far as I can tell, v is always a string.
		// If it's not, tell the user gracefully.
		s, ok := v.(string)
		if !ok {
			return "", fmt.Errorf("unexpected type for variable %q: %T", k, v)
		}
		varmap[k] = ast.Variable{
			Value: s,
			Type:  ast.TypeString,
		}
	}

	cfg := hil.EvalConfig{
		GlobalScope: &ast.BasicScope{
			VarMap:  varmap,
			FuncMap: config.Funcs(),
		},
	}

	result, err := hil.Eval(root, &cfg)
	if err != nil {
		return "", err
	}
	if result.Type != hil.TypeString {
		return "", fmt.Errorf("unexpected output hil.Type: %v", result.Type)
	}

	return result.Value.(string), nil
}

func validateVarsAttribute(v interface{}, key string) (ws []string, es []error) {
	// vars can only be primitives right now
	var badVars []string
	for k, v := range v.(map[string]interface{}) {
		switch v.(type) {
		case []interface{}:
			badVars = append(badVars, fmt.Sprintf("%s (list)", k))
		case map[string]interface{}:
			badVars = append(badVars, fmt.Sprintf("%s (map)", k))
		}
	}
	if len(badVars) > 0 {
		es = append(es, fmt.Errorf(
			"%s: cannot contain non-primitives; bad keys: %s",
			key, strings.Join(badVars, ", ")))
	}
	return
}

func hash(s string) string {
	sha := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sha[:])
}
