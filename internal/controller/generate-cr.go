package controller

import (
	"bytes"
	"text/template"

	sAppv1 "demo.com/pkg/simple/api/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	netv1 "k8s.io/api/networking/v1"
	"k8s.io/apimachinery/pkg/util/yaml"
)

/*
整个逻辑是我当前获取了 crd 对象资源
这个 crd 对象资源中定义了详细的 k8s 资源细节
需要按照细节创建原生 k8s 对象
就需要把数据映射到 yaml 文件格式
*/

// parseTemplate 把 md 中的字段解析到 template 文件中 返回字符数组
func parseTemplate(md *sAppv1.SimpleDemo, templateName string) ([]byte, error) {
	files, err := template.ParseFiles("internal/controller/templates/" + templateName)
	if err != nil {
		return nil, err
	}
	b := &bytes.Buffer{}
	err = files.Execute(b, md)
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// NewDeployment 把 byte 数组按照 yaml 解析到数据
func NewDeployment(md *sAppv1.SimpleDemo) (*appsv1.Deployment, error) {
	content, err := parseTemplate(md, "deployment.yaml")
	if err != nil {
		return nil, err
	}
	deploy := &appsv1.Deployment{}
	//deploy := new(appsv1.Deployment{})
	err = yaml.Unmarshal(content, deploy)
	if err != nil {
		return nil, err
	}
	return deploy, nil
}

// NewIngress 把 byte 数组按照 yaml 解析到数据
func NewIngress(md *sAppv1.SimpleDemo) (*netv1.Ingress, error) {
	content, err := parseTemplate(md, "ingress.yaml")
	if err != nil {
		return nil, err
	}
	ingress := &netv1.Ingress{}
	err = yaml.Unmarshal(content, ingress)
	if err != nil {
		return nil, err
	}
	return ingress, nil
}

// NewService 把 byte 数组按照 yaml 解析到数据
func NewService(md *sAppv1.SimpleDemo) (*corev1.Service, error) {
	content, err := parseTemplate(md, "service.yaml")
	if err != nil {
		return nil, err
	}
	service := &corev1.Service{}
	err = yaml.Unmarshal(content, service)
	if err != nil {
		return nil, err
	}
	return service, nil
}

// NewNodePort 把 byte 数组按照 yaml 解析到数据
func NewNodePort(md *sAppv1.SimpleDemo) (*corev1.Service, error) {
	content, err := parseTemplate(md, "service-np.yaml")
	if err != nil {
		return nil, err
	}
	service := &corev1.Service{}
	err = yaml.Unmarshal(content, service)
	if err != nil {
		return nil, err
	}
	return service, nil
}
