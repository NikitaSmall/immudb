/*
Copyright 2023 Codenotary Inc. All rights reserved.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/codenotary/immudb/embedded/document"
	"github.com/codenotary/immudb/embedded/sql"
	"github.com/codenotary/immudb/pkg/api/authorizationschema"
	"github.com/codenotary/immudb/pkg/api/documentschema"
	"github.com/codenotary/immudb/pkg/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

func TestV2Authentication(t *testing.T) {
	dir := t.TempDir()

	serverOptions := DefaultOptions().
		WithDir(dir).
		WithMetricsServer(false).
		WithAdminPassword(auth.SysAdminPassword).
		WithSigningKey("./../../test/signer/ec1.key")

	s := DefaultServer().WithOptions(serverOptions).(*ImmuServer)

	s.Initialize()

	ctx := context.Background()

	_, err := s.DocumentInsert(ctx, &documentschema.DocumentInsertRequest{})
	assert.ErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.DocumentSearch(ctx, &documentschema.DocumentSearchRequest{})
	assert.ErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.CollectionCreate(ctx, &documentschema.CollectionCreateRequest{})
	assert.ErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.CollectionDelete(ctx, &documentschema.CollectionDeleteRequest{})
	assert.ErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.CollectionList(ctx, &documentschema.CollectionListRequest{})
	assert.ErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.CollectionGet(ctx, &documentschema.CollectionGetRequest{})
	assert.ErrorIs(t, err, ErrNotLoggedIn)

	logged, err := s.OpenSessionV2(ctx, &authorizationschema.OpenSessionRequestV2{
		Username: "immudb",
		Password: "immudb",
		Database: "defaultdb",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, logged.Token)
	fmt.Println(logged.ExpirationTimestamp)
	assert.True(t, logged.InactivityTimestamp > 0)
	assert.True(t, logged.ExpirationTimestamp >= 0)
	assert.True(t, len(logged.ServerUUID) > 0)

	md := metadata.Pairs("sessionid", logged.Token)
	ctx = metadata.NewIncomingContext(context.Background(), md)
	_, err = s.DocumentInsert(ctx, &documentschema.DocumentInsertRequest{})
	assert.NotErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.DocumentSearch(ctx, &documentschema.DocumentSearchRequest{})
	assert.NotErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.CollectionCreate(ctx, &documentschema.CollectionCreateRequest{})
	assert.NotErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.CollectionDelete(ctx, &documentschema.CollectionDeleteRequest{})
	assert.NotErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.CollectionList(ctx, &documentschema.CollectionListRequest{})
	assert.NotErrorIs(t, err, ErrNotLoggedIn)

	_, err = s.CollectionGet(ctx, &documentschema.CollectionGetRequest{})
	assert.NotErrorIs(t, err, ErrNotLoggedIn)

}

func TestPaginationOnReader(t *testing.T) {
	dir := t.TempDir()

	serverOptions := DefaultOptions().
		WithDir(dir).
		WithMetricsServer(false).
		WithAdminPassword(auth.SysAdminPassword).
		WithSigningKey("./../../test/signer/ec1.key")

	s := DefaultServer().WithOptions(serverOptions).(*ImmuServer)
	require.NoError(t, s.Initialize())

	logged, err := s.OpenSessionV2(context.Background(), &authorizationschema.OpenSessionRequestV2{
		Username: "immudb",
		Password: "immudb",
		Database: "defaultdb",
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, logged.Token)

	md := metadata.Pairs("sessionid", logged.Token)
	ctx := metadata.NewIncomingContext(context.Background(), md)

	// create collection
	collectionName := "mycollection"
	_, err = s.CollectionCreate(ctx, &documentschema.CollectionCreateRequest{
		Name: collectionName,
		IndexKeys: map[string]*documentschema.IndexOption{
			"pincode": {Type: documentschema.IndexType_INTEGER},
			"country": {Type: documentschema.IndexType_STRING},
			"idx":     {Type: documentschema.IndexType_INTEGER},
		},
	})
	require.NoError(t, err)

	// add documents to collection
	for i := 1.0; i <= 20; i++ {
		_, err = s.DocumentInsert(ctx, &documentschema.DocumentInsertRequest{
			Collection: collectionName,
			Document: &structpb.Struct{
				Fields: map[string]*structpb.Value{
					"pincode": {
						Kind: &structpb.Value_NumberValue{NumberValue: i},
					},
					"country": {
						Kind: &structpb.Value_StringValue{StringValue: fmt.Sprintf("country-%d", int(i))},
					},
					"idx": {
						Kind: &structpb.Value_NumberValue{NumberValue: i},
					},
				},
			},
		})
		require.NoError(t, err)
	}

	t.Run("test reader for multiple paginated reads", func(t *testing.T) {
		results := make([]*structpb.Struct, 0)

		for i := 1; i <= 4; i++ {
			resp, err := s.DocumentSearch(ctx, &documentschema.DocumentSearchRequest{
				Collection: collectionName,
				Query: []*documentschema.DocumentQuery{
					{
						Field:    "pincode",
						Operator: documentschema.QueryOperator_GE,
						Value: &structpb.Value{
							Kind: &structpb.Value_NumberValue{NumberValue: 0},
						},
					},
				},
				Page:    uint32(i),
				PerPage: 5,
			})
			require.NoError(t, err)
			require.Equal(t, 5, len(resp.Results))
			results = append(results, resp.Results...)
		}

		for i := 1.0; i <= 20; i++ {
			doc := results[int(i-1)]
			data := map[string]interface{}{}
			err = json.Unmarshal([]byte(doc.Fields[document.DocumentBLOBField].GetStringValue()), &data)
			require.NoError(t, err)
			require.Equal(t, i, data["idx"])
		}
	})

	t.Run("test reader should throw no more entries when reading more entries", func(t *testing.T) {
		_, err := s.DocumentSearch(ctx, &documentschema.DocumentSearchRequest{
			Collection: collectionName,
			Query: []*documentschema.DocumentQuery{
				{
					Field:    "pincode",
					Operator: documentschema.QueryOperator_GE,
					Value: &structpb.Value{
						Kind: &structpb.Value_NumberValue{NumberValue: 0},
					},
				},
			},
			Page:    5,
			PerPage: 5,
		})
		require.ErrorIs(t, err, sql.ErrNoMoreRows)

	})

	t.Run("test reader should throw error on reading backwards", func(t *testing.T) {
		for i := 1; i <= 3; i++ {
			resp, err := s.DocumentSearch(ctx, &documentschema.DocumentSearchRequest{
				Collection: collectionName,
				Query: []*documentschema.DocumentQuery{
					{
						Field:    "pincode",
						Operator: documentschema.QueryOperator_GE,
						Value: &structpb.Value{
							Kind: &structpb.Value_NumberValue{NumberValue: 0},
						},
					},
				},
				Page:    uint32(i),
				PerPage: 5,
			})
			require.NoError(t, err)
			require.Equal(t, 5, len(resp.Results))
		}

		_, err := s.DocumentSearch(ctx, &documentschema.DocumentSearchRequest{
			Collection: collectionName,
			Query: []*documentschema.DocumentQuery{
				{
					Field:    "pincode",
					Operator: documentschema.QueryOperator_GE,
					Value: &structpb.Value{
						Kind: &structpb.Value_NumberValue{NumberValue: 0},
					},
				},
			},
			Page:    2, // read upto page 3, check if we can read backwards
			PerPage: 5,
		})

		require.ErrorIs(t, err, ErrInvalidPreviousPage)
	})

}
