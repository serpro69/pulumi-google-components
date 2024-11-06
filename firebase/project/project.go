package project

import (
	"github.com/pulumi/pulumi-gcp/sdk/v8/go/gcp/organizations"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/serpro69/pulumi-google-components/firebase/project/util"
	"github.com/serpro69/pulumi-google-components/firebase/project/vars"
	"github.com/serpro69/pulumi-google-components/project"
)

// FirebaseProject is a struct that represents a project with enabled Firebase support in GCP
type FirebaseProject struct {
	pulumi.ResourceState
	name    string
	Project *organizations.Project
}

// NewProject creates a new Project in GCP
func NewFirebaseProject(
	ctx *pulumi.Context,
	name string,
	args *vars.ProjectArgs,
	opts ...pulumi.ResourceOption,
) (*FirebaseProject, error) {
	p := &FirebaseProject{name: name}
	if err := ctx.RegisterComponentResource(util.Project.String(), name, p, opts...); err != nil {
		return nil, err
	}

	// Required for the project to display in any list of Firebase projects.
	args.Labels = args.Labels.ToStringMapOutput().ApplyT(func(labels map[string]string) map[string]string {
		if l, found := labels["firebase"]; !found || l != "enabled" {
			labels["firebase"] = "enabled"
		}
		return labels
	}).(pulumi.StringMapInput)

	args.ActivateApis = args.ActivateApis.ToStringArrayOutput().ApplyT(func(apis []string) []string {
		apis = append(apis, services...)
		return unique(apis)
	}).(pulumi.StringArrayInput)

	proj, err := project.NewProject(ctx, name, &args.ProjectArgs, pulumi.Parent(p))
	if err != nil {
		return nil, err
	}
	p.Project = proj.Main

	return p, nil
}

func unique[T any](slice []T) []T {
	seen := make(map[any]struct{})
	var result []T
	for _, v := range slice {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

var services = []string{
	// base
	"cloudbilling.googleapis.com",
	"cloudresourcemanager.googleapis.com",
	// By enabling the Service Usage API, the project will be able to accept quota checks!
	// So, for all subsequent resource provisioning and service enabling, you should use the provider with user_project_override (no alias needed).
	"serviceusage.googleapis.com",
	// firebase services
	"firebase.googleapis.com",
	"fcm.googleapis.com",
	"fcmregistrations.googleapis.com",
	"firebaseappdistribution.googleapis.com",
	"firebaseextensions.googleapis.com",
	"firebasedynamiclinks.googleapis.com",
	"firebasehosting.googleapis.com",
	"firebaseinstallations.googleapis.com",
	"firebaseremoteconfig.googleapis.com",
	"firebaseremoteconfigrealtime.googleapis.com",
	"firebaserules.googleapis.com",
	// firebase functions
	// i  functions: ensuring required API cloudfunctions.googleapis.com is enabled...
	// i  functions: ensuring required API cloudbuild.googleapis.com is enabled...
	// i  artifactregistry: ensuring required API artifactregistry.googleapis.com is enabled...
	"cloudfunctions.googleapis.com",
	"cloudbuild.googleapis.com",
	"artifactregistry.googleapis.com",
	// i  functions: packaged .../foo/.firebase/bar/functions (32.51 MB) for uploading
	// i  functions: ensuring required API run.googleapis.com is enabled...
	// i  functions: ensuring required API eventarc.googleapis.com is enabled...
	// i  functions: ensuring required API pubsub.googleapis.com is enabled...
	// i  functions: ensuring required API storage.googleapis.com is enabled...
	"run.googleapis.com",
	"eventarc.googleapis.com",
	"pubsub.googleapis.com",
	"storage.googleapis.com",
}