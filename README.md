# Arangofilters
[LoopBack](http://loopback.io/) inspired filtering system for ArangoDB.

## Overview

Its goal is to provide an easy way of converting JSON filters passed through query strings into an actual AQL query:

```go
// Filter defines a way of filtering AQL queries.
type Filter struct {
  Offset  int                      `json:"offset"`
  Limit   int                      `json:"limit"`
  Sort    []string                 `json:"sort"`
  Where   []map[string]interface{} `json:"where"`
  Options []string                 `json:"options"`
}
```

## Options Field

The `Options` field implementation is left to the developer.
It is not translated into AQL during the filtering.

Its main goal is to allow a filtering similar to the `Include` one in traditional ORMs, as a relation can be a join or a edge in ArangoDB.

Of course, the `Options` field can also be used as a more generic option selector (*e.g.*, `Options: "Basic"` to only return the basic info about a resource).

## Translation example

JSON:
```json
{
  "offset": 1,
  "limit": 2,
  "sort": ["age desc", "money"],
  "where": [
    {"firstName": "Pierre"},
    {
      "or": [
        {"birthPlace": ["Paris", "Los Angeles"]},
        {"age": {"gte": 18}}
      ]
    },
    {
      "like": {
        "text": "lastName",
        "search": "R%",
        "case_insensitive": true
      }
    }
  ]
  },
  "options": ["details"]
}
```

AQL:
```
LIMIT 1, 2
SORT var.age DESC, var.money ASC
FILTER var.'firstName' == 'Pierre' && (var.'birthPlace' IN ['Paris', 'Los Angeles'] || var.'age' >= 18) && LIKE(var.'lastName', 'R%', true)
```

## Operators

- `and`: Logical AND operator.
- `or`: Logical OR operator.
- `not`: Logical NOT operator.
- `gt`, `gte`: Numerical greater than (>); greater than or equal (>=).
- `lt`, `lte`: Numerical less than (<); less than or equal (<=).
- `eq`, `neq`: Equal (==); non equal (!=).
- `like`: LIKE(text, search, case_insensitive) function support

## Usage

Example using [Arangolite](https://github.com/solher/arangolite):

```go
package main

import (
	"context"
	"fmt"

	filters "github.com/solher/arangofilters"
	"github.com/solher/arangolite"
	"github.com/solher/arangolite/requests"
)

func main() {
	ctx := context.Background()

	db := arangolite.NewDatabase(
		arangolite.OptEndpoint("http://localhost:8529"),
		arangolite.OptBasicAuth("user", "password"),
		arangolite.OptDatabaseName("testDB"),
	)

	if err := db.Connect(ctx); err != nil {
		panic(err)
	}

	filter, err := filters.FromJSON(`{"limit": 2}`)
	if err != nil {
		panic(err)
	}

	aqlFilter, err := filters.ToAQL("n", filter)
	if err != nil {
		panic(err)
	}

	r := requests.NewAQL(`
    FOR n
    IN documents
    %s
    RETURN n
  `, aqlFilter)

	documents := []arangolite.Document{}
	if err := db.Run(ctx, r, documents); err != nil {
		panic(err)
	}

	fmt.Println(documents)
}
```
