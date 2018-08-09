package chef_solo_data

import (
	"fmt"
	"strings"
	"testing"

	"encoding/json"
	r "github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"reflect"
)

func TestTemplateChefSoloRendering(t *testing.T) {
	var cases = []struct {
		vars                 string
		default_attributes   string
		automatic_attributes string
		run_list             []string
		policy_name          string
		policy_group         string
		node_id              string
		wantDna              string
		wantNode             string
	}{
		{
			`{}`,
			`{"toto": "po"}`,
			`{"titi": "pi"}`,
			[]string{"a", "b"},
			"",
			"",
			"node_id",
			`{"toto": "po","titi": "pi", "run_list": ["a", "b"], "id": "node_id"}`,
			`{"run_list": ["a", "b"], "default": {"toto": "po"}, "automatic": {"titi": "pi"}, "id": "node_id"}`,
		},
		{
			`{a="foo"}`,
			`{"toto": "$${a}"}`,
			`{"titi": "pi"}`,
			[]string{"a", "b"},
			"",
			"",
			"node_id",
			`{"toto": "foo","titi": "pi", "run_list": ["a", "b"], "id": "node_id"}`,
			`{"run_list": ["a", "b"], "default": {"toto": "foo"}, "automatic": {"titi": "pi"}, "id": "node_id"}`,
		},
		{
			`{a="[\"he\", \"llo\"]"}`,
			`{"toto": $${a}}`,
			`{"titi": "pi"}`,
			[]string{},
			"a",
			"b",
			"node_id",
			`{"toto": ["he", "llo"],"titi": "pi", "policy_name": "a", "policy_group": "b", "id": "node_id"}`,
			`{"policy_name": "a", "policy_group": "b", "default": {"toto": ["he", "llo"]}, "automatic": {"titi": "pi"}, "id": "node_id"}`,
		},
		{
			`{a="[\"he\", \"llo\"]"}`,
			`{"toto": $${a}}`,
			`{"titi": "pi", "fqdn": "$${id}"}`,
			[]string{},
			"a",
			"b",
			"node_id",
			`{"toto": ["he", "llo"],"titi": "pi", "policy_name": "a", "policy_group": "b", "id": "node_id", "fqdn": "node_id"}`,
			`{"policy_name": "a", "policy_group": "b", "default": {"toto": ["he", "llo"]}, "automatic": {"titi": "pi", "fqdn": "node_id"}, "id": "node_id"}`,
		},
	}

	for _, tt := range cases {
		r.UnitTest(t, r.TestCase{
			Providers: testProviders,
			Steps: []r.TestStep{
				r.TestStep{
					Config: testTemplateChefSoloConfig(tt.automatic_attributes, tt.default_attributes,
						tt.run_list, tt.policy_name, tt.policy_group, tt.node_id, tt.vars),
					Check: func(s *terraform.State) error {
						dna := s.RootModule().Outputs["dna"]
						dnaHash := make(map[string]interface{})
						wantedDnaHash := make(map[string]interface{})
						json.Unmarshal([]byte(dna.Value.(string)), &dnaHash)
						json.Unmarshal([]byte(tt.wantDna), &wantedDnaHash)
						if !reflect.DeepEqual(wantedDnaHash, dnaHash) {
							marshalWant, _ := json.Marshal(wantedDnaHash)
							marshalGot, _ := json.Marshal(dnaHash)
							return fmt.Errorf("Error on DNA got:\n%s\nwant:\n%s\n",
								string(marshalGot[:]), string(marshalWant[:]))
						}

						node := s.RootModule().Outputs["node"]
						nodeHash := make(map[string]interface{})
						wantedNodeHash := make(map[string]interface{})
						json.Unmarshal([]byte(node.Value.(string)), &nodeHash)
						json.Unmarshal([]byte(tt.wantNode), &wantedNodeHash)
						if !reflect.DeepEqual(wantedNodeHash, nodeHash) {
							marshalWant, _ := json.Marshal(wantedNodeHash)
							marshalGot, _ := json.Marshal(nodeHash)
							return fmt.Errorf("Error on NODE got:\n%s\nwant:\n%s\n",
								string(marshalGot[:]), string(marshalWant[:]))
						}
						return nil
					},
				},
			},
		})
	}
}

func TestValidateChefSoloVarsAttribute(t *testing.T) {
	cases := map[string]struct {
		Vars      map[string]interface{}
		ExpectErr string
	}{
		"lists are invalid": {
			map[string]interface{}{
				"list": []interface{}{},
			},
			`vars: cannot contain non-primitives`,
		},
		"maps are invalid": {
			map[string]interface{}{
				"map": map[string]interface{}{},
			},
			`vars: cannot contain non-primitives`,
		},
		"strings, integers, floats, and bools are AOK": {
			map[string]interface{}{
				"string": "foo",
				"int":    1,
				"bool":   true,
				"float":  float64(1.0),
			},
			``,
		},
	}

	for tn, tc := range cases {
		_, es := validateVarsAttribute(tc.Vars, "vars")
		if len(es) > 0 {
			if tc.ExpectErr == "" {
				t.Fatalf("%s: expected no err, got: %#v", tn, es)
			}
			if !strings.Contains(es[0].Error(), tc.ExpectErr) {
				t.Fatalf("%s: expected\n%s\nto contain\n%s", tn, es[0], tc.ExpectErr)
			}
		} else if tc.ExpectErr != "" {
			t.Fatalf("%s: expected err containing %q, got none!", tn, tc.ExpectErr)
		}
	}
}

func testTemplateChefSoloConfig(autoAttrs string, defaultAttrs string, runList []string,
	policyName string, policyGroup string, nodeId string, vars string) string {
	data := ""
	if len(runList) > 0 {
		run_list_attr, _ := json.Marshal(runList)

		data = fmt.Sprintf(
			`data "template_chef_solo" "t0" {
					node_id = "%s"
					automatic_attributes = "%s" 
					default_attributes = "%s"
					run_list =  %s
					vars = %s
				}`, nodeId,
			strings.Replace(strings.Replace(autoAttrs, `\`, `\\`, -1), `"`, `\"`, -1),
			strings.Replace(strings.Replace(defaultAttrs, `\`, `\\`, -1), `"`, `\"`, -1),
			string(run_list_attr[:]), vars)
	} else {
		data = fmt.Sprintf(
			`data "template_chef_solo" "t0" {
			node_id = "%s"
			automatic_attributes = "%s" 
			default_attributes = "%s"
            policy_name =  "%s"
			policy_group = "%s"
			vars = %s
		}`, nodeId,
			strings.Replace(strings.Replace(autoAttrs, `\`, `\\`, -1), `"`, `\"`, -1),
			strings.Replace(strings.Replace(defaultAttrs, `\`, `\\`, -1), `"`, `\"`, -1),
			policyName, policyGroup, vars)
	}
	return data + `

		output "dna" {
				value = "${data.template_chef_solo.t0.dna}"
		}
		output "node" {
				value = "${data.template_chef_solo.t0.node}"
		}
		output "environment" {
				value = "${data.template_chef_solo.t0.environment}"
		}
		output "node_id" {
				value = "${data.template_chef_solo.t0.node_id}"
		}
		output "named_run_list" {
				value = "${data.template_chef_solo.t0.named_run_list}"
		}`
}
