package artifactory

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jfrog/jfrog-client-go/artifactory/services"
)

var legacySchema = map[string]*schema.Schema{
	"key": {
		Type:     schema.TypeString,
		Required: true,
		ForceNew: true,
	},
	"package_type": {
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: repoTypeValidator,
	},
	"repositories": {
		Type:     schema.TypeList,
		Elem:     &schema.Schema{Type: schema.TypeString},
		Required: true,
	},
	"description": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"notes": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"includes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
		Default:  "**/*",
	},
	"excludes_pattern": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"repo_layout_ref": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"debian_trivial_layout": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"artifactory_requests_can_retrieve_remote_artifacts": {
		Type:     schema.TypeBool,
		Optional: true,
	},
	"key_pair": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"pom_repository_references_cleanup_policy": {
		Type:     schema.TypeString,
		Optional: true,
		Computed: true,
	},
	"default_deployment_repo": {
		Type:     schema.TypeString,
		Optional: true,
	},
	"force_nuget_authentication": {
		Type:     schema.TypeBool,
		Optional: true,
		Computed: true,
	},
}

var readFunc = mkRepoRead(packVirtualRepository, func() interface{} {
	return &MessyVirtualRepo{}
})

func resourceArtifactoryVirtualRepository() *schema.Resource {
	return &schema.Resource{
		Create: mkRepoCreate(unpackVirtualRepository, readFunc),
		Read:   readFunc,
		Update: mkRepoUpdate(unpackVirtualRepository, readFunc),
		Delete: deleteRepo,
		Exists: repoExists,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},
		Schema: legacySchema,
		DeprecationMessage: "This resource is deprecated and you should use repo type specific resources " +
			"(such as artifactory_virtual_maven_repository) in the future",
	}
}

type MessyVirtualRepo struct {
	services.VirtualRepositoryBaseParams
	services.DebianVirtualRepositoryParams
	services.MavenVirtualRepositoryParams
	services.NugetVirtualRepositoryParams
}

func unpackVirtualRepository(s *schema.ResourceData) (interface{}, string, error) {
	d := &ResourceData{s}
	repo := MessyVirtualRepo{}

	repo.Key = d.getString("key", false)
	repo.Rclass = "virtual"
	repo.PackageType = d.getString("package_type", false)
	repo.IncludesPattern = d.getString("includes_pattern", false)
	repo.ExcludesPattern = d.getString("excludes_pattern", false)
	repo.RepoLayoutRef = d.getString("repo_layout_ref", false)
	repo.DebianTrivialLayout = d.getBoolRef("debian_trivial_layout", false)
	repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts = d.getBoolRef("artifactory_requests_can_retrieve_remote_artifacts", false)
	repo.Repositories = d.getList("repositories")
	repo.Description = d.getString("description", false)
	repo.Notes = d.getString("notes", false)
	repo.KeyPair = d.getString("key_pair", false)
	repo.PomRepositoryReferencesCleanupPolicy = d.getString("pom_repository_references_cleanup_policy", false)
	repo.DefaultDeploymentRepo = d.getString("default_deployment_repo", false)
	// because this doesn't apply to all repo types, RT isn't required to honor what you tell it.
	// So, saying the type is "maven" but then setting this to 'true' doesn't make sense, and RT doesn't seem to care what you tell it
	repo.ForceNugetAuthentication = d.getBoolRef("force_nuget_authentication", false)
	return &repo, repo.Key, nil
}

func packVirtualRepository(r interface{}, d *schema.ResourceData) error {
	repo := r.(*MessyVirtualRepo)
	setValue := mkLens(d)

	setValue("key", repo.Key)
	setValue("package_type", repo.PackageType)
	setValue("description", repo.Description)
	setValue("notes", repo.Notes)
	setValue("includes_pattern", repo.IncludesPattern)
	setValue("excludes_pattern", repo.ExcludesPattern)
	setValue("repo_layout_ref", repo.RepoLayoutRef)
	setValue("debian_trivial_layout", repo.DebianTrivialLayout)
	setValue("artifactory_requests_can_retrieve_remote_artifacts", repo.ArtifactoryRequestsCanRetrieveRemoteArtifacts)
	setValue("key_pair", repo.KeyPair)
	setValue("pom_repository_references_cleanup_policy", repo.PomRepositoryReferencesCleanupPolicy)
	setValue("default_deployment_repo", repo.DefaultDeploymentRepo)
	setValue("repositories", repo.Repositories)
	errors := setValue("force_nuget_authentication", repo.ForceNugetAuthentication)

	if errors != nil && len(errors) > 0 {
		return fmt.Errorf("failed to pack virtual repo %q", errors)
	}

	return nil
}
