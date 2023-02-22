package google_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceGoogleArtifactRegistryRepositoryConfig(t *testing.T) {
	t.Parallel()

	context := map[string]interface{}{
		"random_suffix": acctest.RandString(t, 10),
	}
	funcDataName := "data.google_artifact_registry_repository.my-repo"

	acctest.VcrTest(t, resource.TestCase{
		PreCheck:     func() { acctest.TestAccPreCheck(t) },
		Providers:    acctest.TestAccProviders,
		CheckDestroy: testAccCheckArtifactRegistryRepositoryDestroyProducer(t),
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceGoogleArtifactRegistryRepositoryConfig(context),
				Check: resource.ComposeTestCheckFunc(
					CheckDataSourceStateMatchesResourceState(funcDataName,
						"google_artifact_registry_repository.my-repo"),
				),
			},
		},
	})
}

func testAccDataSourceGoogleArtifactRegistryRepositoryConfig(context map[string]interface{}) string {
	return Nprintf(`
resource "google_artifact_registry_repository" "my-repo" {
  location      = "us-central1"
  repository_id = "tf-test-my-repository%{random_suffix}"
  description   = "example docker repository%{random_suffix}"
  format        = "DOCKER"
}

data "google_artifact_registry_repository" "my-repo" {
  location      = "us-central1"
  repository_id = google_artifact_registry_repository.my-repo.repository_id
}
`, context)
}