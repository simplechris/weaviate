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

package schema

import (
	"context"
	"testing"

	"github.com/pkg/errors"
	"github.com/semi-technologies/weaviate/entities/models"
	"github.com/semi-technologies/weaviate/entities/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// As of now, most class settings are immutable, but we need to allow some
// specific updates, such as the vector index config
func TestClassUpdates(t *testing.T) {
	t.Run("a class which doesn't exist", func(t *testing.T) {
		err := newSchemaManager().UpdateClass(context.Background(),
			nil, "WrongClass", &models.Class{})
		require.NotNil(t, err)
		assert.Equal(t, ErrNotFound, err)
	})

	t.Run("various immutable and mutable fields", func(t *testing.T) {
		type test struct {
			name          string
			initial       *models.Class
			update        *models.Class
			expectedError error
		}

		tests := []test{
			{
				name:    "attempting a name change",
				initial: &models.Class{Class: "InitialName"},
				update:  &models.Class{Class: "UpdatedName"},
				expectedError: errors.Errorf(
					"class name is immutable: " +
						"attempted change from \"InitialName\" to \"UpdatedName\""),
			},
			{
				name:    "attempting to modify the vectorizer",
				initial: &models.Class{Class: "InitialName", Vectorizer: "model1"},
				update:  &models.Class{Class: "InitialName", Vectorizer: "model2"},
				expectedError: errors.Errorf(
					"vectorizer is immutable: " +
						"attempted change from \"model1\" to \"model2\""),
			},
			{
				name:    "attempting to modify the vector index type",
				initial: &models.Class{Class: "InitialName", VectorIndexType: "hnsw"},
				update:  &models.Class{Class: "InitialName", VectorIndexType: "lsh"},
				expectedError: errors.Errorf(
					"vector index type is immutable: " +
						"attempted change from \"hnsw\" to \"lsh\""),
			},
			{
				name:    "attempting to add a property",
				initial: &models.Class{Class: "InitialName"},
				update: &models.Class{
					Class: "InitialName",
					Properties: []*models.Property{
						{
							Name: "newProp",
						},
					},
				},
				expectedError: errors.Errorf(
					"properties cannot be updated through updating the class. Use the add " +
						"property feature (e.g. \"POST /v1/schema/{className}/properties\") " +
						"to add additional properties"),
			},
			{
				name: "leaving properties unchanged",
				initial: &models.Class{
					Class: "InitialName",
					Properties: []*models.Property{
						{
							Name:     "aProp",
							DataType: []string{"string"},
						},
					},
				},
				update: &models.Class{
					Class: "InitialName",
					Properties: []*models.Property{
						{
							Name:     "aProp",
							DataType: []string{"string"},
						},
					},
				},
				expectedError: nil,
			},
			{
				name: "attempting to rename a property",
				initial: &models.Class{
					Class: "InitialName",
					Properties: []*models.Property{
						{
							Name:     "aProp",
							DataType: []string{"string"},
						},
					},
				},
				update: &models.Class{
					Class: "InitialName",
					Properties: []*models.Property{
						{
							Name:     "changedProp",
							DataType: []string{"string"},
						},
					},
				},
				expectedError: errors.Errorf(
					"properties cannot be updated through updating the class. Use the add " +
						"property feature (e.g. \"POST /v1/schema/{className}/properties\") " +
						"to add additional properties"),
			},
			{
				name: "attempting to update the inverted index config",
				initial: &models.Class{
					Class: "InitialName",
					InvertedIndexConfig: &models.InvertedIndexConfig{
						CleanupIntervalSeconds: 17,
					},
				},
				update: &models.Class{
					Class: "InitialName",
					InvertedIndexConfig: &models.InvertedIndexConfig{
						CleanupIntervalSeconds: 18,
					},
				},
				expectedError: errors.Errorf("inverted index config is immutable"),
			},
			{
				name: "attempting to update module config",
				initial: &models.Class{
					Class: "InitialName",
					ModuleConfig: map[string]interface{}{
						"my-module1": map[string]interface{}{
							"my-setting": "some-value",
						},
					},
				},
				update: &models.Class{
					Class: "InitialName",
					ModuleConfig: map[string]interface{}{
						"my-module1": map[string]interface{}{
							"my-setting": "updated-value",
						},
					},
				},
				expectedError: errors.Errorf("module config is immutable"),
			},
			{
				name: "updating vector index config",
				initial: &models.Class{
					Class: "InitialName",
					VectorIndexConfig: map[string]interface{}{
						"some-setting": "old-value",
					},
				},
				update: &models.Class{
					Class: "InitialName",
					VectorIndexConfig: map[string]interface{}{
						"some-setting": "old-value",
					},
				},
				expectedError: nil,
			},
		}

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				sm := newSchemaManager()
				assert.Nil(t, sm.AddObject(context.Background(), nil, test.initial))
				err := sm.UpdateClass(context.Background(), nil, test.initial.Class, test.update)
				if test.expectedError == nil {
					assert.Nil(t, err)
				} else {
					require.NotNil(t, err, "update must error")
					assert.Equal(t, test.expectedError.Error(), err.Error())
				}
			})
		}
	})

	t.Run("update vector index config", func(t *testing.T) {
		t.Run("with a validation error", func(t *testing.T) {
			sm := newSchemaManager()
			migrator := &vectorIndexConfigMigrator{
				validationError: errors.Errorf("don't think so!"),
			}
			sm.migrator = migrator

			t.Run("create an initial class", func(t *testing.T) {
				err := sm.AddObject(context.Background(), nil, &models.Class{
					Class: "ClassWithVectorIndexConfig",
					VectorIndexConfig: map[string]interface{}{
						"setting-1": "value-1",
					},
				})

				assert.Nil(t, err)
			})

			t.Run("attempt an update of the vector index config", func(t *testing.T) {
				err := sm.UpdateClass(context.Background(), nil,
					"ClassWithVectorIndexConfig", &models.Class{
						Class: "ClassWithVectorIndexConfig",
						VectorIndexConfig: map[string]interface{}{
							"setting-1": "updated-value",
						},
					})
				expectedErrMsg := "vector index config: don't think so!"
				expectedValidateCalledWith := fakeVectorConfig{
					raw: map[string]interface{}{
						"setting-1": "updated-value",
					},
				}
				expectedUpdateCalled := false

				require.NotNil(t, err)
				assert.Equal(t, expectedErrMsg, err.Error())
				assert.Equal(t, expectedValidateCalledWith, migrator.validateCalledWith)
				assert.Equal(t, expectedUpdateCalled, migrator.updateCalled)
			})
		})

		t.Run("with a valid update", func(t *testing.T) {
			sm := newSchemaManager()
			migrator := &vectorIndexConfigMigrator{}
			sm.migrator = migrator

			t.Run("create an initial class", func(t *testing.T) {
				err := sm.AddObject(context.Background(), nil, &models.Class{
					Class: "ClassWithVectorIndexConfig",
					VectorIndexConfig: map[string]interface{}{
						"setting-1": "value-1",
					},
				})

				assert.Nil(t, err)
			})

			t.Run("update the vector index config", func(t *testing.T) {
				err := sm.UpdateClass(context.Background(), nil,
					"ClassWithVectorIndexConfig", &models.Class{
						Class: "ClassWithVectorIndexConfig",
						VectorIndexConfig: map[string]interface{}{
							"setting-1": "updated-value",
						},
					})
				expectedValidateCalledWith := fakeVectorConfig{
					raw: map[string]interface{}{
						"setting-1": "updated-value",
					},
				}
				expectedUpdateCalledWith := fakeVectorConfig{
					raw: map[string]interface{}{
						"setting-1": "updated-value",
					},
				}
				expectedUpdateCalled := true

				require.Nil(t, err)
				assert.Equal(t, expectedValidateCalledWith, migrator.validateCalledWith)
				assert.Equal(t, expectedUpdateCalledWith, migrator.updateCalledWith)
				assert.Equal(t, expectedUpdateCalled, migrator.updateCalled)
			})

			t.Run("the update is reflected", func(t *testing.T) {
				class := sm.getClassByName("ClassWithVectorIndexConfig")
				require.NotNil(t, class)
				expectedVectorIndexConfig := fakeVectorConfig{
					raw: map[string]interface{}{
						"setting-1": "updated-value",
					},
				}

				assert.Equal(t, expectedVectorIndexConfig, class.VectorIndexConfig)
			})
		})
	})
}

type vectorIndexConfigMigrator struct {
	NilMigrator
	validationError    error
	validateCalledWith schema.VectorIndexConfig
	updateCalled       bool
	updateCalledWith   schema.VectorIndexConfig
}

func (m *vectorIndexConfigMigrator) ValidateVectorIndexConfigUpdate(ctx context.Context,
	old, updated schema.VectorIndexConfig) error {
	m.validateCalledWith = updated
	return m.validationError
}

func (m *vectorIndexConfigMigrator) UpdateVectorIndexConfig(ctx context.Context,
	className string, updated schema.VectorIndexConfig) error {
	m.updateCalledWith = updated
	m.updateCalled = true
	return nil
}
