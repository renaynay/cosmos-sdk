package ormgraphql

import (
	"github.com/graphql-go/graphql"
)

var pageInfo = graphql.NewObject(graphql.ObjectConfig{
	Name:       "PageInfo",
	Interfaces: nil,
	Fields: map[string]*graphql.Field{
		"hasNextPage": {
			Name:        "hasNextPage",
			Type:        graphql.Boolean,
			Resolve:     nil,
			Subscribe:   nil,
			Description: "",
		},
		"hasPreviousPage": {
			Name:        "hasPreviousPage",
			Type:        graphql.Boolean,
			Resolve:     nil,
			Subscribe:   nil,
			Description: "",
		},
		"startCursor": {
			Name:        "startCursor",
			Type:        cursor,
			Resolve:     nil,
			Subscribe:   nil,
			Description: "",
		},
		"endCursor": {
			Name:        "endCursor",
			Type:        cursor,
			Resolve:     nil,
			Subscribe:   nil,
			Description: "",
		},
	},
	IsTypeOf:    nil,
	Description: "Information about pagination in a connection.",
})
