package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/shurcooL/githubql"
	"golang.org/x/oauth2"
)

func main() {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	)
	httpClient := oauth2.NewClient(context.Background(), src)

	client := githubql.NewClient(httpClient)

	if len(os.Args) < 3 {
		fmt.Fprintln(os.Stderr, "Usage: starcalc <owner> <repo>")
		os.Exit(1)
	}

	owner := os.Args[1]
	repo := os.Args[2]

	var q struct {
		Repository struct {
			Stargazers struct {
				PageInfo struct {
					EndCursor   githubql.String
					HasNextPage githubql.Boolean
				}
				Edges []struct {
					Node struct {
						Login string
					}
					StarredAt string
				}
			} `graphql:"stargazers(first:100,after:$after)"`
		} `graphql:"repository(owner: $owner, name: $repo)"`
	}

	variables := map[string]interface{}{
		"owner": githubql.String(owner),
		"repo":  githubql.String(repo),
		"after": (*githubql.String)(nil),
	}

	for {
		err := client.Query(context.Background(), &q, variables)
		if err != nil {
			panic(err)
		}
		for _, e := range q.Repository.Stargazers.Edges {
			starredAt, _ := time.Parse(time.RFC3339, e.StarredAt)
			fmt.Printf("%s,%s\n", e.Node.Login, starredAt.Format("1/2/2006 15:04:05"))
		}
		if !q.Repository.Stargazers.PageInfo.HasNextPage {
			break
		}
		variables["after"] = githubql.NewString(q.Repository.Stargazers.PageInfo.EndCursor)
	}

}
