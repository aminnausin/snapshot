package helpers

import "github.com/hasura/go-graphql-client"

func StringPtrOrNil(v graphql.String) *string {
	if v == "" {
		return nil
	}
	s := string(v)
	return &s
}
