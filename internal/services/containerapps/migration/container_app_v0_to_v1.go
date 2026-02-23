// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package migration

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-provider-azurerm/internal/tf/pluginsdk"
)

// ContainerAppV0ToV1 migrates container app state from schema version 0 to 1.
//
// The only schema change is converting the `env` blocks inside `template.container`
// and `template.init_container` from TypeList to TypeSet. This fixes cascading diffs
// when env vars are added or removed, because TypeSet compares elements by content
// (hashed by name) instead of by positional index.
//
// See https://github.com/hashicorp/terraform-provider-azurerm/issues/29743
type ContainerAppV0ToV1 struct{}

func (ContainerAppV0ToV1) Schema() map[string]*pluginsdk.Schema {
	// Simplified V0 schema — only structural fields needed for state decoding.
	// Validation functions, descriptions, and defaults are omitted per convention.
	return map[string]*pluginsdk.Schema{
		"name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
		"resource_group_name": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
		"container_app_environment_id": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
		"revision_mode": {
			Type:     pluginsdk.TypeString,
			Required: true,
		},
		"location": {
			Type:     pluginsdk.TypeString,
			Computed: true,
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
					"min_replicas": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
					"max_replicas": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
					"cooldown_period_in_seconds": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
					"polling_interval_in_seconds": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
					"revision_suffix": {
						Type:     pluginsdk.TypeString,
						Optional: true,
						Computed: true,
					},
					"termination_grace_period_seconds": {
						Type:     pluginsdk.TypeInt,
						Optional: true,
					},
					"azure_queue_scale_rule": scaleRuleSchemaV0(),
					"custom_scale_rule":      customScaleRuleSchemaV0(),
					"http_scale_rule":        httpScaleRuleSchemaV0(),
					"tcp_scale_rule":         tcpScaleRuleSchemaV0(),
					"volume":                 volumeSchemaV0(),
				},
			},
		},

		"ingress": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: ingressSchemaV0(),
			},
		},

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

		"dapr": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"app_id":       {Type: pluginsdk.TypeString, Required: true},
					"app_port":     {Type: pluginsdk.TypeInt, Optional: true},
					"app_protocol": {Type: pluginsdk.TypeString, Optional: true},
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

		"workload_profile_name":  {Type: pluginsdk.TypeString, Optional: true},
		"max_inactive_revisions": {Type: pluginsdk.TypeInt, Optional: true},

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
		"latest_revision_name":         {Type: pluginsdk.TypeString, Computed: true},
		"latest_revision_fqdn":         {Type: pluginsdk.TypeString, Computed: true},
		"custom_domain_verification_id": {Type: pluginsdk.TypeString, Computed: true, Sensitive: true},
	}
}

func (ContainerAppV0ToV1) UpgradeFunc() pluginsdk.StateUpgraderFunc {
	return func(ctx context.Context, rawState map[string]interface{}, meta interface{}) (map[string]interface{}, error) {
		// The data representation at the Go level (slices of maps) is identical
		// for TypeList and TypeSet. No transformation of rawState is needed;
		// Terraform re-encodes the state using the new TypeSet schema automatically.
		log.Printf("[DEBUG] Upgrading azurerm_container_app state from v0 to v1 (env TypeList → TypeSet)")
		return rawState, nil
	}
}

// --- V0 helper schemas (env as TypeList) ---

func envSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList, // V0: TypeList (changed to TypeSet in V1)
		Optional: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"name":        {Type: pluginsdk.TypeString, Required: true},
				"value":       {Type: pluginsdk.TypeString, Optional: true},
				"secret_name": {Type: pluginsdk.TypeString, Optional: true},
			},
		},
	}
}

func containerSchemaV0() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name":              {Type: pluginsdk.TypeString, Required: true},
		"image":             {Type: pluginsdk.TypeString, Required: true},
		"cpu":               {Type: pluginsdk.TypeFloat, Required: true},
		"memory":            {Type: pluginsdk.TypeString, Required: true},
		"ephemeral_storage": {Type: pluginsdk.TypeString, Computed: true},
		"env":               envSchemaV0(),
		"args":              {Type: pluginsdk.TypeList, Optional: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
		"command":           {Type: pluginsdk.TypeList, Optional: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
		"liveness_probe":    livenessProbeSchemaV0(),
		"readiness_probe":   readinessProbeSchemaV0(),
		"startup_probe":     startupProbeSchemaV0(),
		"volume_mounts":     volumeMountsSchemaV0(),
	}
}

func initContainerSchemaV0() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"name":              {Type: pluginsdk.TypeString, Required: true},
		"image":             {Type: pluginsdk.TypeString, Required: true},
		"cpu":               {Type: pluginsdk.TypeFloat, Optional: true},
		"memory":            {Type: pluginsdk.TypeString, Optional: true},
		"ephemeral_storage": {Type: pluginsdk.TypeString, Computed: true},
		"env":               envSchemaV0(),
		"args":              {Type: pluginsdk.TypeList, Optional: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
		"command":           {Type: pluginsdk.TypeList, Optional: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
		"volume_mounts":     volumeMountsSchemaV0(),
	}
}

func livenessProbeSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"transport":               {Type: pluginsdk.TypeString, Required: true},
				"port":                    {Type: pluginsdk.TypeInt, Required: true},
				"host":                    {Type: pluginsdk.TypeString, Optional: true},
				"path":                    {Type: pluginsdk.TypeString, Optional: true, Computed: true},
				"initial_delay":           {Type: pluginsdk.TypeInt, Optional: true},
				"interval_seconds":        {Type: pluginsdk.TypeInt, Optional: true},
				"timeout":                 {Type: pluginsdk.TypeInt, Optional: true},
				"failure_count_threshold": {Type: pluginsdk.TypeInt, Optional: true},
				"termination_grace_period_seconds": {Type: pluginsdk.TypeInt, Computed: true},
				"header": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"name":  {Type: pluginsdk.TypeString, Required: true},
							"value": {Type: pluginsdk.TypeString, Required: true},
						},
					},
				},
			},
		},
	}
}

func readinessProbeSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"transport":               {Type: pluginsdk.TypeString, Required: true},
				"port":                    {Type: pluginsdk.TypeInt, Required: true},
				"host":                    {Type: pluginsdk.TypeString, Optional: true},
				"path":                    {Type: pluginsdk.TypeString, Optional: true, Computed: true},
				"initial_delay":           {Type: pluginsdk.TypeInt, Optional: true},
				"interval_seconds":        {Type: pluginsdk.TypeInt, Optional: true},
				"timeout":                 {Type: pluginsdk.TypeInt, Optional: true},
				"failure_count_threshold": {Type: pluginsdk.TypeInt, Optional: true},
				"success_count_threshold": {Type: pluginsdk.TypeInt, Optional: true},
				"header": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"name":  {Type: pluginsdk.TypeString, Required: true},
							"value": {Type: pluginsdk.TypeString, Required: true},
						},
					},
				},
			},
		},
	}
}

func startupProbeSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"transport":               {Type: pluginsdk.TypeString, Required: true},
				"port":                    {Type: pluginsdk.TypeInt, Required: true},
				"host":                    {Type: pluginsdk.TypeString, Optional: true},
				"path":                    {Type: pluginsdk.TypeString, Optional: true, Computed: true},
				"initial_delay":           {Type: pluginsdk.TypeInt, Optional: true},
				"interval_seconds":        {Type: pluginsdk.TypeInt, Optional: true},
				"timeout":                 {Type: pluginsdk.TypeInt, Optional: true},
				"failure_count_threshold": {Type: pluginsdk.TypeInt, Optional: true},
				"termination_grace_period_seconds": {Type: pluginsdk.TypeInt, Computed: true},
				"header": {
					Type:     pluginsdk.TypeList,
					Optional: true,
					Elem: &pluginsdk.Resource{
						Schema: map[string]*pluginsdk.Schema{
							"name":  {Type: pluginsdk.TypeString, Required: true},
							"value": {Type: pluginsdk.TypeString, Required: true},
						},
					},
				},
			},
		},
	}
}

func volumeMountsSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"name":     {Type: pluginsdk.TypeString, Required: true},
				"path":     {Type: pluginsdk.TypeString, Required: true},
				"sub_path": {Type: pluginsdk.TypeString, Optional: true},
			},
		},
	}
}

func volumeSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"name":          {Type: pluginsdk.TypeString, Required: true},
				"storage_type":  {Type: pluginsdk.TypeString, Optional: true},
				"storage_name":  {Type: pluginsdk.TypeString, Optional: true},
				"mount_options": {Type: pluginsdk.TypeString, Optional: true},
			},
		},
	}
}

func scaleRuleSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"name":          {Type: pluginsdk.TypeString, Required: true},
				"queue_name":    {Type: pluginsdk.TypeString, Required: true},
				"queue_length":  {Type: pluginsdk.TypeInt, Required: true},
				"authentication": azureQueueScaleRuleAuthSchemaV0(),
			},
		},
	}
}

func customScaleRuleSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
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
	}
}

// customScaleRuleAuthSchemaV0 is used by both custom_scale_rule (in container_app)
// and event_trigger_config.scale.rules (in container_app_job).
func customScaleRuleAuthSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"secret_name":       {Type: pluginsdk.TypeString, Required: true},
				"trigger_parameter": {Type: pluginsdk.TypeString, Required: true},
			},
		},
	}
}

func httpScaleRuleSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"name":                {Type: pluginsdk.TypeString, Required: true},
				"concurrent_requests": {Type: pluginsdk.TypeString, Required: true},
				"authentication":      httpTcpScaleRuleAuthSchemaV0(),
			},
		},
	}
}

func tcpScaleRuleSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"name":                {Type: pluginsdk.TypeString, Required: true},
				"concurrent_requests": {Type: pluginsdk.TypeString, Required: true},
				"authentication":      httpTcpScaleRuleAuthSchemaV0(),
			},
		},
	}
}

func azureQueueScaleRuleAuthSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Required: true,
		MinItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"secret_name":       {Type: pluginsdk.TypeString, Required: true},
				"trigger_parameter": {Type: pluginsdk.TypeString, Required: true},
			},
		},
	}
}

func httpTcpScaleRuleAuthSchemaV0() *pluginsdk.Schema {
	return &pluginsdk.Schema{
		Type:     pluginsdk.TypeList,
		Optional: true,
		MinItems: 1,
		Elem: &pluginsdk.Resource{
			Schema: map[string]*pluginsdk.Schema{
				"secret_name":       {Type: pluginsdk.TypeString, Required: true},
				"trigger_parameter": {Type: pluginsdk.TypeString, Optional: true},
			},
		},
	}
}

func ingressSchemaV0() map[string]*pluginsdk.Schema {
	return map[string]*pluginsdk.Schema{
		"allow_insecure_connections": {Type: pluginsdk.TypeBool, Optional: true},
		"external_enabled":          {Type: pluginsdk.TypeBool, Optional: true},
		"target_port":               {Type: pluginsdk.TypeInt, Required: true},
		"exposed_port":              {Type: pluginsdk.TypeInt, Optional: true},
		"transport":                 {Type: pluginsdk.TypeString, Optional: true},
		"client_certificate_mode":   {Type: pluginsdk.TypeString, Optional: true},
		"fqdn":                      {Type: pluginsdk.TypeString, Computed: true},
		"custom_domain": {
			Type:     pluginsdk.TypeList,
			Computed: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"certificate_binding_type": {Type: pluginsdk.TypeString, Computed: true},
					"certificate_id":           {Type: pluginsdk.TypeString, Computed: true},
					"name":                     {Type: pluginsdk.TypeString, Computed: true},
				},
			},
		},
		"cors": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			MaxItems: 1,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"allowed_origins":          {Type: pluginsdk.TypeList, Required: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
					"allow_credentials_enabled": {Type: pluginsdk.TypeBool, Optional: true},
					"allowed_headers":          {Type: pluginsdk.TypeList, Optional: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
					"allowed_methods":          {Type: pluginsdk.TypeList, Optional: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
					"exposed_headers":          {Type: pluginsdk.TypeList, Optional: true, Elem: &pluginsdk.Schema{Type: pluginsdk.TypeString}},
					"max_age_in_seconds":       {Type: pluginsdk.TypeInt, Optional: true},
				},
			},
		},
		"ip_security_restriction": {
			Type:     pluginsdk.TypeList,
			Optional: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"action":           {Type: pluginsdk.TypeString, Required: true},
					"ip_address_range": {Type: pluginsdk.TypeString, Required: true},
					"name":             {Type: pluginsdk.TypeString, Required: true},
					"description":      {Type: pluginsdk.TypeString, Optional: true},
				},
			},
		},
		"traffic_weight": {
			Type:     pluginsdk.TypeList,
			Required: true,
			Elem: &pluginsdk.Resource{
				Schema: map[string]*pluginsdk.Schema{
					"label":           {Type: pluginsdk.TypeString, Optional: true},
					"latest_revision": {Type: pluginsdk.TypeBool, Optional: true},
					"revision_suffix": {Type: pluginsdk.TypeString, Optional: true},
					"percentage":      {Type: pluginsdk.TypeInt, Required: true},
				},
			},
		},
	}
}
