package util

import (
	"github.com/cli/go-gh"
	graphql "github.com/cli/shurcooL-graphql"
	"log"
	"strings"
)

type Release struct {
	TagName     string
	Url         string
	IsLatest    bool
	PublishedAt string
}

type ReleaseDetail struct {
	Release
	Description string
}

func parseRepo(repo string) (string, string) {
	parts := strings.SplitN(repo, "/", 2)
	owner := parts[0]
	name := parts[1]

	return owner, name
}

func GetReleases(repo string) []Release {
	owner, name := parseRepo(repo)

	client, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var query struct {
		Repository struct {
			Releases struct {
				Nodes []Release
			} `graphql:"releases(first: $perPage, orderBy: { field: CREATED_AT, direction: DESC })"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":   graphql.String(owner),
		"name":    graphql.String(name),
		"perPage": graphql.Int(30),
	}

	err = client.Query("RepositoryReleases", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return query.Repository.Releases.Nodes
}

func GetReleaseDetail(repo string, tagName string) ReleaseDetail {
	owner, name := parseRepo(repo)

	client, err := gh.GQLClient(nil)
	if err != nil {
		log.Fatal(err)
	}

	var query struct {
		Repository struct {
			Release ReleaseDetail `graphql:"release(tagName: $tagName)"`
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":   graphql.String(owner),
		"name":    graphql.String(name),
		"tagName": graphql.String(tagName),
	}

	err = client.Query("RepositoryRelease", &query, variables)
	if err != nil {
		log.Fatal(err)
	}

	return query.Repository.Release
}
