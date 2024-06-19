package vmwareengine

import (
	"context"
	"log"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-provider-google-beta/google-beta/envvar"
	"github.com/hashicorp/terraform-provider-google-beta/google-beta/sweeper"
	"github.com/hashicorp/terraform-provider-google-beta/google-beta/tpgresource"
	transport_tpg "github.com/hashicorp/terraform-provider-google-beta/google-beta/transport"
)

func init() {
	sweeper.AddTestSweepers("VmwareengineCluster", testSweepVmwareengineCluster)
}

// At the time of writing, the CI only passes us-central1 as the region
func testSweepVmwareengineCluster(region string) error {
	resourceName := "VmwareengineCluster"
	log.Printf("[INFO][SWEEPER_LOG] Starting sweeper for %s", resourceName)

	config, err := sweeper.SharedConfigForRegion(region)
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] error getting shared config for region: %s", err)
		return err
	}

	err = config.LoadAndValidate(context.Background())
	if err != nil {
		log.Printf("[INFO][SWEEPER_LOG] error loading: %s", err)
		return err
	}

	t := &testing.T{}
	billingId := envvar.GetTestBillingAccountFromEnv(t)

	// List of location values includes:
	//   * global location
	//   * regions used for this resource type's acc tests in the past
	//   * the 'region' passed to the sweeper
	locations := []string{region, "global", "southamerica-west1", "me-west1"}
	log.Printf("[INFO][SWEEPER_LOG] Sweeping will include these locations: %v.", locations)
	for _, location := range locations {
		log.Printf("[INFO][SWEEPER_LOG] Beginning the process of sweeping location '%s'.", location)

		// Setup variables to replace in list template
		d := &tpgresource.ResourceDataMock{
			FieldsInSchema: map[string]interface{}{
				"project":         config.Project,
				"region":          location,
				"location":        location,
				"zone":            "-",
				"billing_account": billingId,
			},
		}

		listTemplate := strings.Split("https://vmwareengine.googleapis.com/v1/{{parent}}/clusters", "?")[0]
		listUrl, err := tpgresource.ReplaceVars(d, config, listTemplate)
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] error preparing sweeper list url: %s", err)
			continue
		}

		res, err := transport_tpg.SendRequest(transport_tpg.SendRequestOptions{
			Config:    config,
			Method:    "GET",
			Project:   config.Project,
			RawURL:    listUrl,
			UserAgent: config.UserAgent,
		})
		if err != nil {
			log.Printf("[INFO][SWEEPER_LOG] Error in response from request %s: %s", listUrl, err)
			continue
		}

		resourceList, ok := res["clusters"]
		if !ok {
			log.Printf("[INFO][SWEEPER_LOG] Nothing found in response.")
			continue
		}

		rl := resourceList.([]interface{})

		log.Printf("[INFO][SWEEPER_LOG] Found %d items in %s list response.", len(rl), resourceName)
		// Keep count of items that aren't sweepable for logging.
		nonPrefixCount := 0
		for _, ri := range rl {
			obj := ri.(map[string]interface{})
			if obj["name"] == nil {
				log.Printf("[INFO][SWEEPER_LOG] %s resource name was nil", resourceName)
				continue
			}

			name := tpgresource.GetResourceNameFromSelfLink(obj["name"].(string))
			// Skip resources that shouldn't be sweeped
			if !sweeper.IsSweepableTestResource(name) {
				nonPrefixCount++
				continue
			}

			deleteTemplate := "https://vmwareengine.googleapis.com/v1/{{parent}}/clusters/{{name}}"
			deleteUrl, err := tpgresource.ReplaceVars(d, config, deleteTemplate)
			if err != nil {
				log.Printf("[INFO][SWEEPER_LOG] error preparing delete url: %s", err)
				continue
			}
			deleteUrl = deleteUrl + name

			// Don't wait on operations as we may have a lot to delete
			_, err = transport_tpg.SendRequest(transport_tpg.SendRequestOptions{
				Config:    config,
				Method:    "DELETE",
				Project:   config.Project,
				RawURL:    deleteUrl,
				UserAgent: config.UserAgent,
			})
			if err != nil {
				log.Printf("[INFO][SWEEPER_LOG] Error deleting for url %s : %s", deleteUrl, err)
			} else {
				log.Printf("[INFO][SWEEPER_LOG] Sent delete request for %s resource: %s", resourceName, name)
			}
		}

		if nonPrefixCount > 0 {
			log.Printf("[INFO][SWEEPER_LOG] %d items were non-sweepable and skipped.", nonPrefixCount)
		}
	}
	return nil
}
