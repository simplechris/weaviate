//                           _       _
// __      _____  __ ___   ___  __ _| |_ ___
// \ \ /\ / / _ \/ _` \ \ / / |/ _` | __/ _ \
//  \ V  V /  __/ (_| |\ V /| | (_| | ||  __/
//   \_/\_/ \___|\__,_| \_/ |_|\__,_|\__\___|
//
//  Copyright © 2016 - 2021 SeMI Technologies B.V. All rights reserved.
//
//  CONTACT: hello@semi.technology
//

package nearImage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/semi-technologies/weaviate/entities/modulecapabilities"
	"github.com/semi-technologies/weaviate/entities/moduletools"
)

type Searcher struct {
	vectorizer imgVectorizer
}

func NewSearcher(vectorizer imgVectorizer) *Searcher {
	return &Searcher{vectorizer}
}

type imgVectorizer interface {
	VectorizeImage(ctx context.Context,
		id, image string) ([]float32, error)
}

func (s *Searcher) VectorSearches() map[string]modulecapabilities.VectorForParams {
	vectorSearches := map[string]modulecapabilities.VectorForParams{}
	vectorSearches["nearImage"] = s.vectorForNearImageParam
	return vectorSearches
}

func (s *Searcher) vectorForNearImageParam(ctx context.Context, params interface{},
	findVectorFn modulecapabilities.FindVectorFn,
	cfg moduletools.ClassConfig) ([]float32, error) {
	return s.vectorFromNearImageParam(ctx, params.(*NearImageParams), findVectorFn, cfg)
}

func (s *Searcher) vectorFromNearImageParam(ctx context.Context,
	params *NearImageParams, findVectorFn modulecapabilities.FindVectorFn,
	cfg moduletools.ClassConfig) ([]float32, error) {
	// find vector for given search query
	searchID := fmt.Sprintf("search_%v", time.Now().UnixNano())
	image := s.prepareImage(params.Image)

	vector, err := s.vectorizer.VectorizeImage(ctx, searchID, image)
	if err != nil {
		return nil, errors.Errorf("vectorize image: %v", err)
	}

	return vector, nil
}

func (s *Searcher) prepareImage(image string) string {
	base64 := "base64,"
	indx := strings.LastIndex(image, base64)
	return image[indx+len(base64):]
}
