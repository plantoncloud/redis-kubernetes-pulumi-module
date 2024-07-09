package secret

import (
	"fmt"
	rediskubernetescontextstate "github.com/plantoncloud/redis-kubernetes-pulumi-blueprint/pkg/redis/contextstate"
	"github.com/pulumi/pulumi-random/sdk/v4/go/random"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Resources(ctx *pulumi.Context) (*pulumi.Context, error) {
	randomPassword, err := random.NewRandomPassword(ctx, "generate-admin-password", &random.RandomPasswordArgs{
		Length:     pulumi.Int(12),
		Special:    pulumi.Bool(true),
		Numeric:    pulumi.Bool(true),
		Upper:      pulumi.Bool(true),
		Lower:      pulumi.Bool(true),
		MinSpecial: pulumi.Int(3),
		MinNumeric: pulumi.Int(2),
		MinUpper:   pulumi.Int(2),
		MinLower:   pulumi.Int(2),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get random password value: %w", err)
	}

	var ctxConfig = ctx.Value(rediskubernetescontextstate.Key).(rediskubernetescontextstate.ContextState)

	addRandomPasswordToContext(&ctxConfig, randomPassword)
	ctx = ctx.WithValue(rediskubernetescontextstate.Key, ctxConfig)
	return ctx, nil
}

func addRandomPasswordToContext(existingConfig *rediskubernetescontextstate.ContextState, randomPassword *random.RandomPassword) {
	if existingConfig.Status.AddedResources == nil {
		existingConfig.Status.AddedResources = &rediskubernetescontextstate.AddedResources{
			RandomPassword: randomPassword,
		}
		return
	}
	existingConfig.Status.AddedResources.RandomPassword = randomPassword
}
