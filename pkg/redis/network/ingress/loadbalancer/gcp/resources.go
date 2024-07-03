package gcp

import (
	"github.com/pkg/errors"
	rediscontextstate "github.com/plantoncloud/redis-kubernetes-pulumi-blueprint/pkg/redis/contextstate"
	redisloadbalancercommon "github.com/plantoncloud/redis-kubernetes-pulumi-blueprint/pkg/redis/network/ingress/loadbalancer/common"
	pulumikubernetescorev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) (*pulumi.Context, error) {
	// Create a Kubernetes Service of type LoadBalancer
	externalLoadBalancerService, err := addExternal(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add external load balancer")
	}
	internalLoadBalancerService, err := addInternal(ctx)
	if err != nil {
		return nil, errors.Wrap(err, "failed to add internal load balancer")
	}

	var contextState = ctx.Value(rediscontextstate.Key).(rediscontextstate.ContextState)

	addLoadBalancerExternalServiceToContext(&contextState, externalLoadBalancerService)
	addLoadBalancerInternalServiceToContext(&contextState, internalLoadBalancerService)
	ctx = ctx.WithValue(rediscontextstate.Key, contextState)

	return ctx, nil
}

func addExternal(ctx *pulumi.Context) (*pulumikubernetescorev1.Service, error) {
	i := extractInput(ctx)
	addedKubeService, err := pulumikubernetescorev1.NewService(ctx,
		redisloadbalancercommon.ExternalLoadBalancerServiceName,
		getLoadBalancerServiceArgs(i, redisloadbalancercommon.ExternalLoadBalancerServiceName, i.externalHostname),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "30s", Update: "30s", Delete: "30s"}), pulumi.Parent(i.namespace))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kubernetes service of type load balancer")
	}
	return addedKubeService, nil
}

func addInternal(ctx *pulumi.Context) (*pulumikubernetescorev1.Service, error) {
	i := extractInput(ctx)
	addedKubeService, err := pulumikubernetescorev1.NewService(ctx,
		redisloadbalancercommon.InternalLoadBalancerServiceName,
		getInternalLoadBalancerServiceArgs(i, i.internalHostname, i.namespace),
		pulumi.Timeouts(&pulumi.CustomTimeouts{Create: "30s", Update: "30s", Delete: "30s"}), pulumi.Parent(i.namespace))
	if err != nil {
		return nil, errors.Wrap(err, "failed to create kubernetes service of type load balancer")
	}
	return addedKubeService, nil
}

func getInternalLoadBalancerServiceArgs(i *input, hostname string, namespace *pulumikubernetescorev1.Namespace) *pulumikubernetescorev1.ServiceArgs {
	resp := getLoadBalancerServiceArgs(i, redisloadbalancercommon.InternalLoadBalancerServiceName, hostname)
	resp.Metadata = &metav1.ObjectMetaArgs{
		Name:      pulumi.String(redisloadbalancercommon.InternalLoadBalancerServiceName),
		Namespace: namespace.Metadata.Name(),
		Labels:    namespace.Metadata.Labels(),
		Annotations: pulumi.StringMap{
			"cloud.google.com/load-balancer-type":       pulumi.String("Internal"),
			"planton.cloud/endpoint-domain-name":        pulumi.String(i.endpointDomainName),
			"external-dns.alpha.kubernetes.io/hostname": pulumi.String(hostname),
		},
	}
	return resp
}

func getLoadBalancerServiceArgs(i *input, serviceName, hostname string) *pulumikubernetescorev1.ServiceArgs {
	return &pulumikubernetescorev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String(serviceName),
			Namespace: i.namespace.Metadata.Name(),
			Annotations: pulumi.StringMap{
				"planton.cloud/endpoint-domain-name":        pulumi.String(i.endpointDomainName),
				"external-dns.alpha.kubernetes.io/hostname": pulumi.String(hostname)}},
		Spec: &pulumikubernetescorev1.ServiceSpecArgs{
			Type: pulumi.String("LoadBalancer"), // Service type is LoadBalancer
			Ports: pulumikubernetescorev1.ServicePortArray{
				&pulumikubernetescorev1.ServicePortArgs{
					Name:       pulumi.String("tcp-redis"),
					Port:       pulumi.Int(6379),
					Protocol:   pulumi.String("TCP"),
					TargetPort: pulumi.String("redis"), // This assumes your Redis pod has a port named 'redis'
				},
			},
			Selector: pulumi.StringMap{
				"app.kubernetes.io/component": pulumi.String("master"),
				"app.kubernetes.io/instance":  pulumi.String("redis"),
				"app.kubernetes.io/name":      pulumi.String("redis"),
			},
		},
	}
}

func addLoadBalancerExternalServiceToContext(existingConfig *rediscontextstate.ContextState, loadBalancerService *pulumikubernetescorev1.Service) {
	if existingConfig.Status.AddedResources == nil {
		existingConfig.Status.AddedResources = &rediscontextstate.AddedResources{
			LoadBalancerExternalService: loadBalancerService,
		}
		return
	}
	existingConfig.Status.AddedResources.LoadBalancerExternalService = loadBalancerService
}

func addLoadBalancerInternalServiceToContext(existingConfig *rediscontextstate.ContextState, loadBalancerService *pulumikubernetescorev1.Service) {
	if existingConfig.Status.AddedResources == nil {
		existingConfig.Status.AddedResources = &rediscontextstate.AddedResources{
			LoadBalancerInternalService: loadBalancerService,
		}
		return
	}
	existingConfig.Status.AddedResources.LoadBalancerInternalService = loadBalancerService
}
