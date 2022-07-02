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

func GetReleases(repo string) []Release {
	parts := strings.SplitN(repo, "/", 2)
	owner := parts[0]
	name := parts[1]

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
