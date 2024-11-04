/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// SimpleDemoSpec defines the desired state of SimpleDemo.
type SimpleDemoSpec struct {
	// Image 镜像地址
	Image string `json:"image"`
	// Port 服务提供的端口
	Port int32 `json:"port"`
	// Replicas 要部署多少个副本
	// +optional
	Replicas int32 `json:"replicas,omitempty"`
	// StartCmd 启动命令
	// +optional
	StartCmd string `json:"startCmd,omitempty"`
	// Args 启动命令参数
	// +optional
	Args []string `json:"args,omitempty"`
	// Environments 环境变量 直接使用 pod 中的定义方式
	// +optional
	Environments []corev1.EnvVar `json:"environments,omitempty"`
	// Expose service 要暴露的端口
	Expose *Expose `json:"expose"`
}

// Expose defines the desired state of Expose.
type Expose struct {
	// Mode 模式 nodePort or ingress
	Mode string `json:"mode"`
	// IngressDomain 域名 在 Mode 为 ingress 时 必填
	// +optional
	IngressDomain string `json:"ingressDomain,omitempty"`
	// NodePort nodePort 端口 在 Mode 为 nodePort 时 必填
	// +optional
	NodePort int32 `json:"nodePort,omitempty"`
	// ServicePort service 端口 一般随机生成 这里使用和服务相同的端口 Port
	// +optional
	ServicePort int32 `json:"servicePort,omitempty"`
}

// SimpleDemoStatus defines the observed state of SimpleDemo.
type SimpleDemoStatus struct {
	// Phase 处于什么阶段
	// +optional
	Phase string `json:"phase,omitempty"`
	// Message 这个阶段信息
	// +optional
	Message string `json:"message,omitempty"`
	// Reason 处于这个阶段的原因
	// +optional
	Reason string `json:"reason,omitempty"`
	// Conditions 处于这个阶段的原因
	// +optional
	Conditions []Condition `json:"conditions,omitempty"`
}

// Condition defines the observed state of Condition.
type Condition struct {
	// Type 子资源类型
	// +optional
	Type string `json:"type,omitempty"`
	// Message 子资源状态信息
	// +optional
	Message string `json:"message,omitempty"`
	// Status 子资源状态名称
	// +optional
	Status string `json:"status,omitempty"`
	// Reason 子资源状态原因
	// +optional
	Reason string `json:"reason,omitempty"`
	// LastTransitionTime 最后 创建/更新 时间
	// +optional
	LastTransitionTime metav1.Time `json:"lastTransitionTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// SimpleDemo is the Schema for the simpledemoes API.
type SimpleDemo struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SimpleDemoSpec   `json:"spec,omitempty"`
	Status SimpleDemoStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SimpleDemoList contains a list of SimpleDemo.
type SimpleDemoList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SimpleDemo `json:"items"`
}

func init() {
	SchemeBuilder.Register(&SimpleDemo{}, &SimpleDemoList{})
}
