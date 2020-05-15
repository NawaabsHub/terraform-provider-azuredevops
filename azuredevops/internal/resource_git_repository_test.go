// +build all core resource_git_repository

package internal

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/microsoft/azure-devops-go-api/azuredevops/core"
	"github.com/microsoft/azure-devops-go-api/azuredevops/git"
	"github.com/microsoft/terraform-provider-azuredevops/azdosdkmocks"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/config"
	"github.com/microsoft/terraform-provider-azuredevops/azuredevops/utils/converter"
	"github.com/stretchr/testify/require"
)

var testRepoProjectID = uuid.New()
var testRepoID = uuid.New()

// This definition matches the overall structure of what a configured git repository would
// look like. Note that the ID and Name attributes match -- this is the service-side behavior
// when configuring a GitHub repo.
var testAzureGitRepository = git.GitRepository{
	Id:   &testRepoID,
	Name: converter.String("RepoName"),
	Project: &core.TeamProjectReference{
		Id:   &testRepoProjectID,
		Name: converter.String("ProjectName"),
	},
}

/**
 * Begin unit tests
 */

// verifies that the create operation is considered failed if the initial API
// call fails.
func TestAzureGitRepo_Create_DoesNotSwallowErrorFromFailedCreateCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceData := schema.TestResourceDataRaw(t, ResourceGitRepository().Schema, nil)
	resourceData.SetId(testAzureGitRepository.Id.String())
	flattenGitRepository(resourceData, &testAzureGitRepository)
	configureCleanInitialization(resourceData)

	reposClient := azdosdkmocks.NewMockGitClient(ctrl)
	clients := &config.AggregatedClient{GitReposClient: reposClient, Ctx: context.Background()}

	expectedArgs := git.CreateRepositoryArgs{
		GitRepositoryToCreate: &git.GitRepositoryCreateOptions{
			Name: testAzureGitRepository.Name,
			Project: &core.TeamProjectReference{
				Id: &testRepoProjectID,
			},
		},
	}
	reposClient.
		EXPECT().
		CreateRepository(clients.Ctx, expectedArgs).
		Return(nil, errors.New("CreateAzureGitRepository() Failed")).
		Times(1)

	err := resourceGitRepositoryCreate(resourceData, clients)
	require.Regexp(t, ".*CreateAzureGitRepository\\(\\) Failed$", err.Error())
}

// verifies that the update operation is considered failed if the initial API
// call fails.
func TestAzureGitRepo_Update_DoesNotSwallowErrorFromFailedCreateCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	resourceData := schema.TestResourceDataRaw(t, ResourceGitRepository().Schema, nil)
	resourceData.SetId(testAzureGitRepository.Id.String())
	flattenGitRepository(resourceData, &testAzureGitRepository)
	configureCleanInitialization(resourceData)

	reposClient := azdosdkmocks.NewMockGitClient(ctrl)
	clients := &config.AggregatedClient{GitReposClient: reposClient, Ctx: context.Background()}

	reposClient.
		EXPECT().
		UpdateRepository(clients.Ctx, gomock.Any()).
		Return(nil, errors.New("UpdateAzureGitRepository() Failed")).
		Times(1)

	err := resourceGitRepositoryUpdate(resourceData, clients)
	require.Regexp(t, ".*UpdateAzureGitRepository\\(\\) Failed$", err.Error())
}

func configureCleanInitialization(d *schema.ResourceData) {
	d.Set("initialization", &[]map[string]interface{}{
		{
			"init_type": "Clean",
		},
	})
}

// verifies that a round-trip flatten/expand sequence will not result in data loss of non-computed properties.
//	Note: there is no need to expand computed properties, so they won't be tested here.
func TestAzureGitRepo_FlattenExpand_RoundTrip(t *testing.T) {
	projectID := uuid.New()
	project := core.TeamProjectReference{Id: &projectID}

	repoID := uuid.New()
	repoName := "name"
	gitRepo := git.GitRepository{Id: &repoID, Name: &repoName, Project: &project}

	resourceData := schema.TestResourceDataRaw(t, ResourceGitRepository().Schema, nil)
	resourceData.SetId(gitRepo.Id.String())
	flattenGitRepository(resourceData, &gitRepo)

	expandedGitRepo, repoInitialization, expandedProjectID, err := expandGitRepository(resourceData)

	require.Nil(t, err)
	require.NotNil(t, expandedGitRepo)
	require.NotNil(t, expandedGitRepo.Id)
	require.Equal(t, *expandedGitRepo.Id, repoID)
	require.NotNil(t, expandedProjectID)
	require.Equal(t, *expandedProjectID, projectID)
	require.Nil(t, repoInitialization)
}

func TestAzureGitRepo_FlattenExpandInitialization_RoundTrip(t *testing.T) {
	projectID := uuid.New()
	project := core.TeamProjectReference{Id: &projectID}

	repoID := uuid.New()
	repoName := "name"
	gitRepo := git.GitRepository{Id: &repoID, Name: &repoName, Project: &project}

	resourceData := schema.TestResourceDataRaw(t, ResourceGitRepository().Schema, nil)
	resourceData.SetId(gitRepo.Id.String())
	flattenGitRepository(resourceData, &gitRepo)
	configureCleanInitialization(resourceData)

	expandedGitRepo, repoInitialization, expandedProjectID, err := expandGitRepository(resourceData)

	require.Nil(t, err)
	require.NotNil(t, expandedGitRepo)
	require.NotNil(t, expandedGitRepo.Id)
	require.Equal(t, *expandedGitRepo.Id, repoID)
	require.NotNil(t, expandedProjectID)
	require.Equal(t, *expandedProjectID, projectID)
	require.NotNil(t, repoInitialization)
	require.Equal(t, repoInitialization.initType, "Clean")
	require.Equal(t, repoInitialization.sourceType, "")
	require.Equal(t, repoInitialization.sourceURL, "")
}

// verifies that the read operation is considered failed if the initial API
// call fails.
func TestAzureGitRepo_Read_DoesNotSwallowErrorFromFailedReadCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reposClient := azdosdkmocks.NewMockGitClient(ctrl)
	clients := &config.AggregatedClient{
		GitReposClient: reposClient,
		Ctx:            context.Background(),
	}

	resourceData := schema.TestResourceDataRaw(t, ResourceGitRepository().Schema, nil)
	resourceData.SetId("an-id")
	resourceData.Set("project_id", "a-project")

	expectedArgs := git.GetRepositoryArgs{RepositoryId: converter.String("an-id"), Project: converter.String("a-project")}
	reposClient.
		EXPECT().
		GetRepository(clients.Ctx, expectedArgs).
		Return(nil, fmt.Errorf("GetRepository() Failed")).
		Times(1)

	err := resourceGitRepositoryRead(resourceData, clients)
	require.Contains(t, err.Error(), "GetRepository() Failed")
}

// verifies that the resource ID is used for reads if the ID is set
func TestAzureGitRepo_Read_UsesIdIfSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reposClient := azdosdkmocks.NewMockGitClient(ctrl)
	clients := &config.AggregatedClient{
		GitReposClient: reposClient,
		Ctx:            context.Background(),
	}

	resourceData := schema.TestResourceDataRaw(t, ResourceGitRepository().Schema, nil)
	resourceData.SetId("an-id")
	resourceData.Set("project_id", "a-project")

	expectedArgs := git.GetRepositoryArgs{RepositoryId: converter.String("an-id"), Project: converter.String("a-project")}
	reposClient.
		EXPECT().
		GetRepository(clients.Ctx, expectedArgs).
		Return(nil, fmt.Errorf("error")).
		Times(1)

	resourceGitRepositoryRead(resourceData, clients)
}

func TestAzureGitRepo_Delete_ChecksForValidUUID(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, ResourceGitRepository().Schema, nil)
	resourceData.SetId("not-a-uuid-id")

	err := resourceGitRepositoryDelete(resourceData, &config.AggregatedClient{})
	require.NotNil(t, err)
	require.Contains(t, err.Error(), "Invalid repositoryId UUID")
}

func TestAzureGitRepo_Delete_DoesNotSwallowErrorFromFailedDeleteCall(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reposClient := azdosdkmocks.NewMockGitClient(ctrl)
	clients := &config.AggregatedClient{
		GitReposClient: reposClient,
		Ctx:            context.Background(),
	}

	resourceData := schema.TestResourceDataRaw(t, ResourceGitRepository().Schema, nil)
	id := uuid.New()
	resourceData.SetId(id.String())

	expectedArgs := git.DeleteRepositoryArgs{RepositoryId: &id}
	reposClient.
		EXPECT().
		DeleteRepository(clients.Ctx, expectedArgs).
		Return(fmt.Errorf("DeleteRepository() Failed")).
		Times(1)

	err := resourceGitRepositoryDelete(resourceData, clients)
	require.Contains(t, err.Error(), "DeleteRepository() Failed")
}

// verifies that the name is used for reads if the ID is not set
func TestAzureGitRepo_Read_UsesNameIfIdNotSet(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	reposClient := azdosdkmocks.NewMockGitClient(ctrl)
	clients := &config.AggregatedClient{
		GitReposClient: reposClient,
		Ctx:            context.Background(),
	}

	resourceData := schema.TestResourceDataRaw(t, ResourceGitRepository().Schema, nil)
	resourceData.Set("name", "a-name")
	resourceData.Set("project_id", "a-project")

	expectedArgs := git.GetRepositoryArgs{RepositoryId: converter.String("a-name"), Project: converter.String("a-project")}
	reposClient.
		EXPECT().
		GetRepository(clients.Ctx, expectedArgs).
		Return(nil, fmt.Errorf("error")).
		Times(1)

	resourceGitRepositoryRead(resourceData, clients)
}
