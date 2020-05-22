// GENERATED BY THE COMMAND ABOVE; DO NOT EDIT
// This file was generated by swaggo/swag at
// 2020-05-22 19:11:06.513755 +0800 CST m=+0.951733069

package docs

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/alecthomas/template"
	"github.com/swaggo/swag"
)

var doc = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{.Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "license": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/api/common/kinds": {
            "get": {
                "description": "Get all chaos kinds from Kubernetes cluster.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "common"
                ],
                "summary": "Get all chaos kinds from Kubernetes cluster.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    }
                }
            }
        },
        "/api/common/namespaces": {
            "get": {
                "description": "Get all namespaces from Kubernetes cluster.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "common"
                ],
                "summary": "Get all namespaces from Kubernetes cluster.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "type": "string"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    }
                }
            }
        },
        "/api/common/pods": {
            "get": {
                "description": "Get pods from Kubernetes cluster.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "common"
                ],
                "summary": "Get pods from Kubernetes cluster.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "namespace",
                        "name": "namespace",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/common.Pod"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    }
                }
            }
        },
        "/api/experiments": {
            "get": {
                "description": "Get chaos experiments from Kubernetes cluster.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "experiments"
                ],
                "summary": "Get chaos experiments from Kubernetes cluster.",
                "parameters": [
                    {
                        "type": "string",
                        "description": "namespace",
                        "name": "namespace",
                        "in": "query"
                    },
                    {
                        "type": "string",
                        "description": "name",
                        "name": "name",
                        "in": "query"
                    },
                    {
                        "enum": [
                            "PodChaos",
                            "IoChaos",
                            "NetworkChaos",
                            "TimeChaos",
                            "KernelChaos",
                            "StressChaos"
                        ],
                        "type": "string",
                        "description": "kind",
                        "name": "kind",
                        "in": "query"
                    },
                    {
                        "enum": [
                            "Running",
                            "Paused",
                            "Failed",
                            "Finished"
                        ],
                        "type": "string",
                        "description": "status",
                        "name": "status",
                        "in": "query"
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/experiment.Experiment"
                            }
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    }
                }
            }
        },
        "/api/experiments/new": {
            "post": {
                "description": "Create a new chaos experiments.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "experiments"
                ],
                "summary": "Create a nex chaos experiments.",
                "parameters": [
                    {
                        "description": "Request body",
                        "name": "request",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/experiment.ExperimentInfo"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "create ok"
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    }
                }
            }
        },
        "/api/experiments/pause/{kind}/{ns}/{name}": {
            "put": {
                "description": "Pause chaos experiment by API",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "experiments"
                ],
                "summary": "Pause chaos experiment by API",
                "parameters": [
                    {
                        "type": "string",
                        "description": "kind",
                        "name": "kind",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "namespace",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "name",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "pause ok"
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    }
                }
            }
        },
        "/api/experiments/start/{kind}/{ns}/{name}": {
            "put": {
                "description": "Start the paused chaos experiment by API",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "experiments"
                ],
                "summary": "Start the paused chaos experiment by API",
                "parameters": [
                    {
                        "type": "string",
                        "description": "kind",
                        "name": "kind",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "namespace",
                        "name": "namespace",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "name",
                        "name": "name",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "start ok"
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    }
                }
            }
        },
        "/api/experiments/state": {
            "get": {
                "description": "Get chaos experiments state from Kubernetes cluster.",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "experiments"
                ],
                "summary": "Get chaos experiments state from Kubernetes cluster.",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/experiment.ChaosState"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/utils.APIError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "common.Pod": {
            "type": "object",
            "properties": {
                "ip": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "namespace": {
                    "type": "string"
                }
            }
        },
        "experiment.ChaosState": {
            "type": "object",
            "properties": {
                "failed": {
                    "type": "integer"
                },
                "finished": {
                    "type": "integer"
                },
                "paused": {
                    "type": "integer"
                },
                "running": {
                    "type": "integer"
                },
                "total": {
                    "type": "integer"
                }
            }
        },
        "experiment.Experiment": {
            "type": "object",
            "required": [
                "kind",
                "name",
                "namespace"
            ],
            "properties": {
                "created": {
                    "type": "string"
                },
                "kind": {
                    "type": "string"
                },
                "name": {
                    "type": "string"
                },
                "namespace": {
                    "type": "string"
                },
                "status": {
                    "type": "string"
                }
            }
        },
        "experiment.ExperimentInfo": {
            "type": "object",
            "required": [
                "name",
                "namespace"
            ],
            "properties": {
                "annotations": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "labels": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "name": {
                    "type": "string"
                },
                "namespace": {
                    "type": "string"
                },
                "scheduler": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.SchedulerInfo"
                },
                "scope": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.ScopeInfo"
                },
                "target": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.TargetInfo"
                }
            }
        },
        "experiment.IOChaosInfo": {
            "type": "object"
        },
        "experiment.KernelChaosInfo": {
            "type": "object"
        },
        "experiment.NetworkChaosInfo": {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string"
                },
                "bandwidth": {
                    "type": "object",
                    "$ref": "#/definitions/v1alpha1.BandwidthSpec"
                },
                "corrupt": {
                    "type": "object",
                    "$ref": "#/definitions/v1alpha1.CorruptSpec"
                },
                "delay": {
                    "type": "object",
                    "$ref": "#/definitions/v1alpha1.DelaySpec"
                },
                "direction": {
                    "type": "string"
                },
                "duplicate": {
                    "type": "object",
                    "$ref": "#/definitions/v1alpha1.DuplicateSpec"
                },
                "loss": {
                    "type": "object",
                    "$ref": "#/definitions/v1alpha1.LossSpec"
                },
                "target_scope": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.ScopeInfo"
                }
            }
        },
        "experiment.PodChaosInfo": {
            "type": "object",
            "properties": {
                "action": {
                    "type": "string"
                },
                "container_name": {
                    "type": "string"
                }
            }
        },
        "experiment.SchedulerInfo": {
            "type": "object",
            "properties": {
                "cron": {
                    "type": "string"
                },
                "duration": {
                    "type": "string"
                }
            }
        },
        "experiment.ScopeInfo": {
            "type": "object",
            "properties": {
                "annotation_selectors": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "field_selectors": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "label_selectors": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "mode": {
                    "type": "string"
                },
                "namespace_selectors": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "phase_selectors": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "value": {
                    "type": "string"
                }
            }
        },
        "experiment.StressChaosInfo": {
            "type": "object"
        },
        "experiment.TargetInfo": {
            "type": "object",
            "required": [
                "kind"
            ],
            "properties": {
                "io_chaos": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.IOChaosInfo"
                },
                "kernel_chaos": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.KernelChaosInfo"
                },
                "kind": {
                    "type": "string"
                },
                "network_chaos": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.NetworkChaosInfo"
                },
                "pod_chaos": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.PodChaosInfo"
                },
                "stress_chaos": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.StressChaosInfo"
                },
                "time_chaos": {
                    "type": "object",
                    "$ref": "#/definitions/experiment.TimeChaosInfo"
                }
            }
        },
        "experiment.TimeChaosInfo": {
            "type": "object"
        },
        "utils.APIError": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "string"
                },
                "error": {
                    "type": "boolean"
                },
                "full_text": {
                    "type": "string"
                },
                "message": {
                    "type": "string"
                }
            }
        },
        "v1alpha1.BandwidthSpec": {
            "type": "object",
            "properties": {
                "buffer": {
                    "description": "Buffer is the maximum amount of bytes that tokens can be available for instantaneously.\n+kubebuilder:validation:Minimum=1",
                    "type": "integer"
                },
                "limit": {
                    "description": "Limit is the number of bytes that can be queued waiting for tokens to become available.\n+kubebuilder:validation:Minimum=1",
                    "type": "integer"
                },
                "minburst": {
                    "description": "Minburst specifies the size of the peakrate bucket. For perfect\naccuracy, should be set to the MTU of the interface.  If a\npeakrate is needed, but some burstiness is acceptable, this\nsize can be raised. A 3000 byte minburst allows around 3mbit/s\nof peakrate, given 1000 byte packets.\n+optional\n+kubebuilder:validation:Minimum=0",
                    "type": "integer"
                },
                "peakrate": {
                    "description": "Peakrate is the maximum depletion rate of the bucket.\nThe peakrate does not need to be set, it is only necessary\nif perfect millisecond timescale shaping is required.\n+optional\n+kubebuilder:validation:Minimum=0",
                    "type": "integer"
                },
                "rate": {
                    "description": "Rate is the speed knob. Allows bps, kbps, mbps, gbps, tbps unit. bps means bytes per second.",
                    "type": "string"
                }
            }
        },
        "v1alpha1.CorruptSpec": {
            "type": "object",
            "properties": {
                "correlation": {
                    "type": "string"
                },
                "corrupt": {
                    "type": "string"
                }
            }
        },
        "v1alpha1.DelaySpec": {
            "type": "object",
            "properties": {
                "correlation": {
                    "type": "string"
                },
                "jitter": {
                    "type": "string"
                },
                "latency": {
                    "type": "string"
                },
                "reorder": {
                    "type": "object",
                    "$ref": "#/definitions/v1alpha1.ReorderSpec"
                }
            }
        },
        "v1alpha1.DuplicateSpec": {
            "type": "object",
            "properties": {
                "correlation": {
                    "type": "string"
                },
                "duplicate": {
                    "type": "string"
                }
            }
        },
        "v1alpha1.LossSpec": {
            "type": "object",
            "properties": {
                "correlation": {
                    "type": "string"
                },
                "loss": {
                    "type": "string"
                }
            }
        },
        "v1alpha1.ReorderSpec": {
            "type": "object",
            "properties": {
                "correlation": {
                    "type": "string"
                },
                "gap": {
                    "type": "integer"
                },
                "reorder": {
                    "type": "string"
                }
            }
        }
    }
}`

type swaggerInfo struct {
	Version     string
	Host        string
	BasePath    string
	Schemes     []string
	Title       string
	Description string
}

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = swaggerInfo{
	Version:     "",
	Host:        "",
	BasePath:    "",
	Schemes:     []string{},
	Title:       "",
	Description: "",
}

type s struct{}

func (s *s) ReadDoc() string {
	sInfo := SwaggerInfo
	sInfo.Description = strings.Replace(sInfo.Description, "\n", "\\n", -1)

	t, err := template.New("swagger_info").Funcs(template.FuncMap{
		"marshal": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}).Parse(doc)
	if err != nil {
		return doc
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, sInfo); err != nil {
		return doc
	}

	return tpl.String()
}

func init() {
	swag.Register(swag.Name, &s{})
}
