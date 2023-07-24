// Copyright (c) 2018-2023, Sylabs Inc. All rights reserved.
// This software is licensed under a 3-clause BSD license. Please consult the
// LICENSE.md file distributed with the sources of this project regarding your
// rights to use or distribute this software.

package client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
)

var errQueryValueMustBeSpecified = errors.New("search query ('value') must be specified")

// Search performs a library search, returning any matching collections,
// containers, entities, or images.
//
// args specifies key-value pairs to be used as a search spec, such as "arch"
// (ie. "amd64") or "signed" (valid values "true" or "false").
//
// "value" is a required keyword for all searches. It will be matched against
// all collections (Entity, Collection, Container, and Image)
//
// Multiple architectures may be searched by specifying a comma-separated list
// (ie. "amd64,arm64") for the value of "arch".
//
// Match all collections with name "thename":
//
//	c.Search(ctx, map[string]string{"value": "thename"})
//
// Match all images with name "imagename" and arch "amd64"
//
//	c.Search(ctx, map[string]string{
//	    "value": "imagename",
//	    "arch": "amd64"
//	})
//
// Note: if 'arch' and/or 'signed' are specified, the search is limited in
// scope only to the "Image" collection.
func (c *Client) Search(ctx context.Context, args map[string]string) (*SearchResults, error) {
	// "value" is minimally required in "args"
	value, ok := args["value"]
	if !ok {
		return nil, errQueryValueMustBeSpecified
	}

	if len(value) < 3 {
		return nil, fmt.Errorf("%w: bad query '%s'. You must search for at least 3 characters", errBadRequest, value)
	}

	v := url.Values{}
	for key, value := range args {
		v.Set(key, value)
	}

	resJSON, err := c.apiGet(ctx, "v1/search?"+v.Encode())
	if err != nil {
		return nil, err
	}

	var res SearchResponse
	if err := json.Unmarshal(resJSON, &res); err != nil {
		return nil, fmt.Errorf("error decoding results: %w", err)
	}

	return &res.Data, nil
}
