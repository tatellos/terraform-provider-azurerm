// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package migration

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

// ContainerAppJobV0ToV1 migrates container app job state from schema version 0 to 1.
//
// The only schema change is converting the `env` blocks inside
// `template.container` and `template.init_container` from TypeList to TypeSet.
//
// See https://github.com/hashicorp/terraform-provider-azurerm/issues/29743
type ContainerAppJobV0ToV1 struct{}

func (ContainerAppJobV0ToV1) Schema() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
		"resource_group_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
		"location": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
		"container_app_environment_id": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
		"replica_timeout_in_seconds": {
			Type:     pluginsdk.TypeInt,
			Required: true,
		},
		"workload_profile_name": {
			Type:     pluginsdk.TypeString,
			Optional: true,
		},

		"template": {
			Type:     pluginsdk.TypeList,
			Required: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"container": {
						Type:     pluginsdk.TypeList,
						Required: true,
						MinItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: containerSchemaV0(),
						},
					},
					"init_container": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						MinItems: 1,
						Elem: &pluginsdk.Resource{
							Schema: initContainerSchemaV0(),
						},
					},
					"volume": volumeSchemaV0(),
				},
			},
		},

		"secret": {
			Type:      pluginsdk.TypeSet,
			Optional:  true,
			Sensitive: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"name":                {Type: pluginsdk.TypeString, Required: true},
					"value":               {Type: pluginsdk.TypeString, Optional: true, Sensitive: true},
					"key_vault_secret_id": {Type: pluginsdk.TypeString, Optional: true},
					"identity":            {Type: pluginsdk.TypeString, Optional: true},
				},
			},
		},

		"replica_retry_limit": {Type: pluginsdk.TypeInt, Optional: true},

		"registry": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MinItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"server":               {Type: pluginsdk.TypeString, Required: true},
					"username":             {Type: pluginsdk.TypeString, Optional: true},
					"password_secret_name": {Type: pluginsdk.TypeString, Optional: true},
					"identity":             {Type: pluginsdk.TypeString, Optional: true},
				},
			},
		},

		"event_trigger_config": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"parallelism":            {Type: pluginsdk.TypeInt, Optional: true},
					"replica_completion_count": {Type: pluginsdk.TypeInt, Optional: true},
					"scale": {
						Type:     pluginsdk.TypeList,
						Optional: true,
						Elem: &pluginsdk.Resource{
							Schema: map[string]*pluginsdk.Schema{
								"max_executions":              {Type: pluginsdk.TypeInt, Optional: true},
								"min_executions":              {Type: pluginsdk.TypeInt, Optional: true},
								"polling_interval_in_seconds": {Type: pluginsdk.TypeInt, Optional: true},
								"rules": {
									Type:     pluginsdk.TypeList,
									Optional: true,
									Elem: &pluginsdk.Resource{
										Schema: map[string]*pluginsdk.Schema{
											"name":             {Type: pluginsdk.TypeString, Required: true},
											"custom_rule_type": {Type: pluginsdk.TypeString, Required: true},
											"metadata":         {Type: pluginsdk.TypeMap, Required: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
											"authentication":   customScaleRuleAuthSchemaV0(),
										},
									},
								},
							},
						},
					},
				},
			},
		},

		"schedule_trigger_config": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"cron_expression":        {Type: pluginsdk.TypeString, Required: true},
					"parallelism":            {Type: pluginsdk.TypeInt, Optional: true},
					"replica_completion_count": {Type: pluginsdk.TypeInt, Optional: true},
				},
			},
		},

		"manual_trigger_config": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"parallelism":            {Type: pluginsdk.TypeInt, Optional: true},
					"replica_completion_count": {Type: pluginsdk.TypeInt, Optional: true},
				},
			},
		},

		"identity": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"type":         {Type: pluginsdk.TypeString, Required: true},
					"identity_ids": {Type: pluginsdk.TypeSet, Optional: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
					"principal_id": {Type: pluginsdk.TypeString, Computed: true},
					"tenant_id":    {Type: pluginsdk.TypeString, Computed: true},
				},
			},
		},

		"tags": {
			Type:     pluginsdk.TypeMap,
			Optional: true,
			Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
		},

		"outbound_ip_addresses": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem:     &pluginsdk.Schema{Type: pluginsdk.TypeString},
		},
		"event_stream_endpoint": {Type: pluginsdk.TypeString, Computed: true},
	}
}

func (ContainerAppJobV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		log.Printf("[DEBUG] Upgrading azurerm_container_app_job state from v0 to v1 (env TypeList → TypeSet)")
		return rawState, nil
	}
}
