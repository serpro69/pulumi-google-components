package webapps

import (
	"errors"
	"fmt"

	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/firebase"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/serpro69/pulumi-google-components/firebase/webapps/util"
	"github.com/serpro69/pulumi-google-components/firebase/webapps/vars"
)

type FirebaseProjectWebApps struct {
	pulumi.ResourceState

	Apps    firebase.WebAppArrayOutput
	Configs []*firebase.GetWebAppConfigResult
	Domains firebase.HostingCustomDomainArray
}

func ConfigureWebApps(
	ctx *pulumi.Context,
	name string,
	args *vars.ProjectWebAppsArgs,
	opts ...pulumi.ResourceOption,
) (*FirebaseProjectWebApps, error) {
	// Check for mandatory arguments
	if args == nil || args.ProjectId == nil {
		return nil, errors.New("ProjectId is mandatory")
	}

	fbWebApps := &FirebaseProjectWebApps{}
	if err := ctx.RegisterComponentResource(util.WebApps.String(), name, fbWebApps, opts...); err != nil {
		return nil, err
	}

	wa := args.ProjectId.ToStringOutput().ApplyT(func(projectId string) (firebase.WebAppArrayOutput, error) {
		webApps := args.WebApps.ToStringArrayOutput().ApplyT(func(apps []string) ([]*firebase.WebApp, error) {
			var aa []*firebase.WebApp
			for _, app := range apps {
				a, err := firebase.NewWebApp(ctx, app,
					&firebase.WebAppArgs{
						Project:     pulumi.String(projectId),
						DisplayName: pulumi.String(app),
					},
					pulumi.Parent(fbWebApps),
				)
				if err != nil {
					return nil, err
				}
				aa = append(aa, a)

				a.AppId.ApplyT(func(appId string) error {
					ac, err := firebase.GetWebAppConfig(ctx,
						&firebase.GetWebAppConfigArgs{
							Project:  pulumi.StringRef(projectId),
							WebAppId: appId,
						},
						pulumi.Parent(a),
					)
					if err != nil {
						return err
					}
					fbWebApps.Configs = append(fbWebApps.Configs, ac)
					return nil
				})

				hs, err := firebase.NewHostingSite(ctx, app,
					&firebase.HostingSiteArgs{
						Project: a.Project,
						AppId:   a.AppId,
						SiteId:  pulumi.Sprintf("%s-%s", app, projectId),
					},
					pulumi.Parent(a),
				)
				if err != nil {
					return nil, err
				}

				args.CustomDomains.ToStringArrayMapOutput().ApplyT(func(domains map[string][]string) error {
					for _, domain := range domains[app] {
						hs.SiteId.ApplyT(func(siteId *string) error {
							d, err := firebase.NewHostingCustomDomain(ctx, fmt.Sprintf("%s$%s", app, domain),
								&firebase.HostingCustomDomainArgs{
									Project:        a.Project,
									SiteId:         pulumi.String(*siteId),
									CertPreference: pulumi.String("DEDICATED"),
									CustomDomain:   pulumi.String(domain),
								},
								pulumi.Parent(hs),
							)
							if err != nil {
								return err
							}
							fbWebApps.Domains = append(fbWebApps.Domains, d)
							return nil
						})
					}
					return nil
				})
			}
			return aa, nil
		}).(firebase.WebAppArrayOutput)
		return webApps, nil
	}).(firebase.WebAppArrayOutput)

	fbWebApps.Apps = wa

	if err := ctx.RegisterResourceOutputs(fbWebApps, pulumi.Map{
		"apps":    wa,
		"configs": pulumi.All(fbWebApps.Configs).ApplyT(func(configs []interface{}) []interface{} { return configs }),
		"domains": fbWebApps.Domains,
	}); err != nil {
		return nil, err
	}

	return fbWebApps, nil
}
