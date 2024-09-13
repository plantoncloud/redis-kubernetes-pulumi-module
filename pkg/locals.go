package pkg

import (
	"fmt"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/code2cloud/v1/kubernetes/rediskubernetes"
	"github.com/plantoncloud/planton-cloud-apis/zzgo/cloud/planton/apis/commons/apiresource/enums/apiresourcekind"
	"github.com/plantoncloud/pulumi-module-golang-commons/pkg/provider/kubernetes/kuberneteslabelkeys"
	"github.com/plantoncloud/redis-kubernetes-pulumi-module/pkg/outputs"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"strconv"
)

type Locals struct {
	IngressExternalHostname string
	IngressInternalHostname string
	KubePortForwardCommand  string
	KubeServiceFqdn         string
	KubeServiceName         string
	Namespace               string
	RedisKubernetes         *rediskubernetes.RedisKubernetes
	RedisPodSelectorLabels  map[string]string
	Labels                  map[string]string
}

func initializeLocals(ctx *pulumi.Context, stackInput *rediskubernetes.RedisKubernetesStackInput) *Locals {
	locals := &Locals{}

	redisKubernetes := stackInput.ApiResource

	//assign value for the local variable to make it available across the module.
	locals.RedisKubernetes = redisKubernetes

	locals.Labels = map[string]string{
		kuberneteslabelkeys.Environment:  stackInput.ApiResource.Spec.EnvironmentInfo.EnvId,
		kuberneteslabelkeys.Organization: stackInput.ApiResource.Spec.EnvironmentInfo.OrgId,
		kuberneteslabelkeys.Resource:     strconv.FormatBool(true),
		kuberneteslabelkeys.ResourceId:   stackInput.ApiResource.Metadata.Id,
		kuberneteslabelkeys.ResourceKind: apiresourcekind.ApiResourceKind_redis_kubernetes.String(),
	}

	//decide on the namespace
	locals.Namespace = redisKubernetes.Metadata.Id

	ctx.Export(outputs.Namespace, pulumi.String(locals.Namespace))

	locals.RedisPodSelectorLabels = map[string]string{
		"app.kubernetes.io/component": "master",
		"app.kubernetes.io/instance":  redisKubernetes.Metadata.Id,
		"app.kubernetes.io/name":      "redis",
	}

	locals.KubeServiceName = fmt.Sprintf("%s-master", redisKubernetes.Metadata.Name)

	//export kubernetes service name
	ctx.Export(outputs.Service, pulumi.String(locals.KubeServiceName))

	locals.KubeServiceFqdn = fmt.Sprintf("%s.%s.svc.cluster.local", locals.KubeServiceName, locals.Namespace)

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
