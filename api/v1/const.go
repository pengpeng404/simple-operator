package v1

const (
	ModeIngress  = "ingress"
	ModeNodePort = "nodeport"
)

const (
	ConditionTypeDeployment = "Deployment"
	ConditionTypeService    = "Service"
	ConditionTypeIngress    = "Ingress"

	ConditionMessageDeploymentOKFmt  = "Deployment %s is ready"
	ConditionMessageDeploymentNOTFmt = "Deployment %s is not ready"
	ConditionMessageServiceOKFmt     = "Service %s is ready"
	ConditionMessageServiceNOTFmt    = "Service %s is not ready"
	ConditionMessageIngressOKFmt     = "Ingress %s is ready"
	ConditionMessageIngressNOTFmt    = "Ingress %s is not ready"

	ConditionReasonDeploymentReady    = "Deployment Ready"
	ConditionReasonDeploymentNOTReady = "Deployment NOT Ready"
	ConditionReasonServiceReady       = "Service Ready"
	ConditionReasonServiceNOTReady    = "Service NOT Ready"
	ConditionReasonIngressReady       = "Ingress Ready"
	ConditionReasonIngressNOTReady    = "Ingress NOT Ready"

	ConditionStatusTrue  = "True"
	ConditionStatusFalse = "False"
)

const (
	StatusReasonSuccess  = "Success"
	StatusMessageSuccess = "Success"
	StatusPhaserComplete = "Complete"
)
