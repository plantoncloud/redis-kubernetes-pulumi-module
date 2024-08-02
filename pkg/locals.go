package pkg

import (
	"fmt"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/rediskubernetes/model"
	"github.com/plantoncloud/redis-kubernetes-pulumi-module/pkg/outputs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type Locals struct {
	IngressExternalHostname string
	IngressInternalHostname string
	KubePortForwardCommand  string
	KubeServiceFqdn         string
	KubeServiceName         string
	Namespace               string
	RedisKubernetes         *model.RedisKubernetes
	RedisPodSelectorLabels  map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *model.RedisKubernetesStackInput) *Locals {
	locals := &Locals{}

	redisKubernetes := stackInput.ApiResource

	//assign value for the local variable to make it available across the module.
	locals.RedisKubernetes = redisKubernetes

	//decide on the namespace
	locals.Namespace = redisKubernetes.Metadata.Id

	locals.RedisPodSelectorLabels = map[string]string{
		"app.kubernetes.io/component": "master",
		"app.kubernetes.io/instance":  redisKubernetes.Metadata.Id,
		"app.kubernetes.io/name":      "redis",
	}

	locals.KubeServiceName = fmt.Sprintf("%s-master", redisKubernetes.Metadata.Name)

	//export kubernetes service name
	ctx.Export(outputs.Service, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local.", locals.KubeServiceName, locals.Namespace)

	//export kubernetes endpoint
	ctx.Export(outputs.KubeEndpoint, pulumi.String(locals.KubeServiceFqdn))

	locals.KubePortForwardCommand = fmt.Sprintf("kubectl port-forward -n %s service/%s 8080:8080",
		locals.Namespace, locals.KubeServiceName)

	//export kube-port-forward command
	ctx.Export(outputs.KubePortForwardCommand, pulumi.String(locals.KubePortForwardCommand))

	if redisKubernetes.Spec.Ingress == nil ||
		!redisKubernetes.Spec.Ingress.IsEnabled ||
		redisKubernetes.Spec.Ingress.EndpointDomainName == "" {
		return locals
	}

	locals.IngressExternalHostname = fmt.Sprintf("%s.%s", redisKubernetes.Metadata.Id,
		redisKubernetes.Spec.Ingress.EndpointDomainName)

	locals.IngressInternalHostname = fmt.Sprintf("%s-internal.%s", redisKubernetes.Metadata.Id,
		redisKubernetes.Spec.Ingress.EndpointDomainName)

	//export ingress hostnames
	ctx.Export(outputs.IngressExternalHostname, pulumi.String(locals.IngressExternalHostname))
	ctx.Export(outputs.IngressInternalHostname, pulumi.String(locals.IngressInternalHostname))

	return locals
}
