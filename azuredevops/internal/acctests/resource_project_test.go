// +build all core resource_project

package azuredevops

// The tests in this file use the mock clients in mock_client.go to mock out
// the Azure DevOps client operations.

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/testhelper"
)

// Verifies that the following sequence of events occurrs without error:
//	(1) TF apply creates project
//	(2) TF state values are set
//	(3) project can be queried by ID and has expected name
//  (4) TF apply update project with changing name
//  (5) project can be queried by ID and has expected name
// 	(6) TF destroy deletes project
//	(7) project can no longer be queried by ID
func TestAccAzureDevOpsProject_CreateAndUpdate(t *testing.T) {
	projectNameFirst := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	projectNameSecond := testhelper.TestAccResourcePrefix + acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum)
	tfNode := "azuredevops_project.project"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testhelper.TestAccPreCheck(t, nil) },
		Providers:    TestProviders(),
		CheckDestroy: testAccProjectCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testhelper.TestAccProjectResource(projectNameFirst),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "project_name", projectNameFirst),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "private"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
					testAccCheckProjectResourceExists(projectNameFirst),
				),
			},
			{
				Config: testhelper.TestAccProjectResource(projectNameSecond),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(tfNode, "process_template_id"),
					resource.TestCheckResourceAttr(tfNode, "project_name", projectNameSecond),
					resource.TestCheckResourceAttr(tfNode, "version_control", "Git"),
					resource.TestCheckResourceAttr(tfNode, "visibility", "private"),
					resource.TestCheckResourceAttr(tfNode, "work_item_template", "Agile"),
					testAccCheckProjectResourceExists(projectNameSecond),
				),
			},
			{
				// Resource Acceptance Testing https://www.terraform.io/docs/extend/resources/import.html#resource-acceptance-testing-implementation
				ResourceName:      tfNode,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// Given the name of an AzDO project, this will return a function that will check whether
// or not the project (1) exists in the state and (2) exist in AzDO and (3) has the correct name
func testAccCheckProjectResourceExists(expectedName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources["azuredevops_project.project"]
		if !ok {
			return fmt.Errorf("Did not find a project in the TF state")
		}

		clients := TestProvider().Meta().(*config.AggregatedClient)
		id := resource.Primary.ID
		project, err := ProjectRead(clients, id, "")

		if err != nil {
			return fmt.Errorf("Project with ID=%s cannot be found!. Error=%v", id, err)
		}

		if *project.Name != expectedName {
			return fmt.Errorf("Project with ID=%s has Name=%s, but expected Name=%s", id, *project.Name, expectedName)
		}

		return nil
	}
}

// verifies that all projects referenced in the state are destroyed. This will be invoked
// *after* terrafform destroys the resource but *before* the state is wiped clean.
func testAccProjectCheckDestroy(s *terraform.State) error {
	clients := TestProvider().Meta().(*config.AggregatedClient)

	// verify that every project referenced in the state does not exist in AzDO
	for _, resource := range s.RootModule().Resources {
		if resource.Type != "azuredevops_project" {
			continue
		}

		id := resource.Primary.ID

		// indicates the project still exists - this should fail the test
		if _, err := ProjectRead(clients, id, ""); err == nil {
			return fmt.Errorf("project with ID %s should not exist", id)
		}
	}

	return nil
}
