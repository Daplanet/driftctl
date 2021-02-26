package github

import (
	"context"
	"testing"

	"github.com/cloudskiff/driftctl/mocks"
	"github.com/pkg/errors"
	"github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListRepositoriesForUser_WithError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{},
	}

	_, err := r.ListRepositories()
	assert.Equal(expectedError, err)
}

func TestListRepositoriesForUser(t *testing.T) {

	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOwnerQuery)
			if !ok {
				return false
			}
			q.Viewer.Repositories.Nodes = []struct{ Name string }{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Viewer.Repositories.PageInfo = pageInfo{
				EndCursor:   "next",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"cursor": (*githubv4.String)(nil),
		}).Return(nil)

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOwnerQuery)
			if !ok {
				return false
			}
			q.Viewer.Repositories.Nodes = []struct{ Name string }{
				{
					Name: "repo3",
				},
				{
					Name: "repo4",
				},
			}
			q.Viewer.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"cursor": githubv4.NewString("next"),
		}).Return(nil)

	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{},
	}

	repos, err := r.ListRepositories()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal([]string{
		"repo1",
		"repo2",
		"repo3",
		"repo4",
	}, repos)
}

func TestListRepositoriesForOrganization_WithError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
	}

	_, err := r.ListRepositories()
	assert.Equal(expectedError, err)
}

func TestListRepositoriesForOrganization(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				EndCursor:   "next",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil)

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo3",
				},
				{
					Name: "repo4",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("testorg"),
			"cursor": githubv4.NewString("next"),
		}).Return(nil)

	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
	}

	repos, err := r.ListRepositories()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal([]string{
		"repo1",
		"repo2",
		"repo3",
		"repo4",
	}, repos)
}

func TestListTeams_WithError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
	}

	_, err := r.ListTeams()
	assert.Equal(expectedError, err)
}

func TestListTeams_WithoutOrganization(t *testing.T) {
	assert := assert.New(t)

	r := githubRepository{}

	teams, err := r.ListTeams()
	assert.Nil(err)
	assert.Equal([]int{}, teams)
}

func TestListTeams(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listTeamsQuery)
			if !ok {
				return false
			}
			q.Organization.Teams.Nodes = []struct {
				DatabaseId int
			}{
				{
					DatabaseId: 1,
				},
				{
					DatabaseId: 2,
				},
			}
			q.Organization.Teams.PageInfo = pageInfo{
				EndCursor:   "next",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil)

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listTeamsQuery)
			if !ok {
				return false
			}
			q.Organization.Teams.Nodes = []struct {
				DatabaseId int
			}{
				{
					DatabaseId: 3,
				},
				{
					DatabaseId: 4,
				},
			}
			q.Organization.Teams.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"login":  (githubv4.String)("testorg"),
			"cursor": githubv4.NewString("next"),
		}).Return(nil)

	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "testorg",
		},
	}

	teams, err := r.ListTeams()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal([]int{1, 2, 3, 4}, teams)
}

func TestListBranchProtection_WithRepoListingError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("my-organization"),
			"cursor": (*githubv4.String)(nil),
		}).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "my-organization",
		},
	}

	_, err := r.ListBranchProtection()
	assert.Equal(expectedError, err)
}

func TestListBranchProtection_WithError(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}
	expectedError := errors.New("test error from graphql")
	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("testorg"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil)

	mockedClient.On("Query", mock.Anything, mock.Anything, mock.Anything).Return(expectedError)

	r := githubRepository{
		client: &mockedClient,
		config: githubConfig{
			Organization: "testorg",
		},
	}

	_, err := r.ListBranchProtection()
	assert.Equal(expectedError, err)
}

func TestListBranchProtection(t *testing.T) {
	assert := assert.New(t)

	mockedClient := mocks.GithubGraphQLClient{}

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listRepoForOrgQuery)
			if !ok {
				return false
			}
			q.Organization.Repositories.Nodes = []struct {
				Name string
			}{
				{
					Name: "repo1",
				},
				{
					Name: "repo2",
				},
			}
			q.Organization.Repositories.PageInfo = pageInfo{
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"org":    (githubv4.String)("my-organization"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil)

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listBranchProtectionQuery)
			if !ok {
				return false
			}
			q.Repository.BranchProtectionRules.Nodes = []struct {
				Id string
			}{
				{
					Id: "id1",
				},
				{
					Id: "id2",
				},
			}
			q.Repository.BranchProtectionRules.PageInfo = pageInfo{
				EndCursor:   "nextPage",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"owner":  (githubv4.String)("my-organization"),
			"name":   (githubv4.String)("repo1"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil)

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listBranchProtectionQuery)
			if !ok {
				return false
			}
			q.Repository.BranchProtectionRules.Nodes = []struct {
				Id string
			}{
				{
					Id: "id3",
				},
				{
					Id: "id4",
				},
			}
			q.Repository.BranchProtectionRules.PageInfo = pageInfo{
				EndCursor:   "nextPage",
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"owner":  (githubv4.String)("my-organization"),
			"name":   (githubv4.String)("repo1"),
			"cursor": (githubv4.String)("nextPage"),
		}).Return(nil)

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listBranchProtectionQuery)
			if !ok {
				return false
			}
			q.Repository.BranchProtectionRules.Nodes = []struct {
				Id string
			}{
				{
					Id: "id5",
				},
				{
					Id: "id6",
				},
			}
			q.Repository.BranchProtectionRules.PageInfo = pageInfo{
				EndCursor:   "nextPage",
				HasNextPage: true,
			}
			return true
		}),
		map[string]interface{}{
			"owner":  (githubv4.String)("my-organization"),
			"name":   (githubv4.String)("repo2"),
			"cursor": (*githubv4.String)(nil),
		}).Return(nil)

	mockedClient.On("Query",
		mock.Anything,
		mock.MatchedBy(func(query interface{}) bool {
			q, ok := query.(*listBranchProtectionQuery)
			if !ok {
				return false
			}
			q.Repository.BranchProtectionRules.Nodes = []struct {
				Id string
			}{
				{
					Id: "id7",
				},
				{
					Id: "id8",
				},
			}
			q.Repository.BranchProtectionRules.PageInfo = pageInfo{
				EndCursor:   "nextPage",
				HasNextPage: false,
			}
			return true
		}),
		map[string]interface{}{
			"owner":  (githubv4.String)("my-organization"),
			"name":   (githubv4.String)("repo2"),
			"cursor": (githubv4.String)("nextPage"),
		}).Return(nil)

	r := githubRepository{
		client: &mockedClient,
		ctx:    context.TODO(),
		config: githubConfig{
			Organization: "my-organization",
		},
	}

	teams, err := r.ListBranchProtection()
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal([]string{
		"id1",
		"id2",
		"id3",
		"id4",
		"id5",
		"id6",
		"id7",
		"id8",
	}, teams)
}
