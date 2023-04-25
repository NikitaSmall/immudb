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

	"github.com/codenotary/immudb/pkg/api/documentschema"
	"github.com/codenotary/immudb/pkg/api/schema"
)

func (s *ImmuServer) DocumentInsert(ctx context.Context, req *documentschema.DocumentInsertRequest) (*documentschema.DocumentInsertResponse, error) {
	db, err := s.getDBFromCtx(ctx, "DocumentInsert")
	if err != nil {
		return nil, err
	}
	resp, err := db.InsertDocument(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ImmuServer) DocumentUpdate(ctx context.Context, req *documentschema.DocumentUpdateRequest) (*documentschema.DocumentUpdateResponse, error) {
	db, err := s.getDBFromCtx(ctx, "DocumentUpdate")
	if err != nil {
		return nil, err
	}
	resp, err := db.UpdateDocument(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ImmuServer) DocumentSearch(ctx context.Context, req *documentschema.DocumentSearchRequest) (*documentschema.DocumentSearchResponse, error) {
	db, err := s.getDBFromCtx(ctx, "DocumentSearch")
	if err != nil {
		return nil, err
	}
	resp, err := db.SearchDocuments(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ImmuServer) CollectionCreate(ctx context.Context, req *documentschema.CollectionCreateRequest) (*documentschema.CollectionCreateResponse, error) {
	db, err := s.getDBFromCtx(ctx, "CollectionCreate")
	if err != nil {
		return nil, err
	}
	resp, err := db.CreateCollection(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ImmuServer) CollectionGet(ctx context.Context, req *documentschema.CollectionGetRequest) (*documentschema.CollectionGetResponse, error) {
	db, err := s.getDBFromCtx(ctx, "CollectionGet")
	if err != nil {
		return nil, err
	}
	resp, err := db.GetCollection(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ImmuServer) CollectionList(ctx context.Context, req *documentschema.CollectionListRequest) (*documentschema.CollectionListResponse, error) {
	db, err := s.getDBFromCtx(ctx, "CollectionList")
	if err != nil {
		return nil, err
	}
	resp, err := db.ListCollections(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// TODO: implement
func (s *ImmuServer) CollectionDelete(ctx context.Context, req *documentschema.CollectionDeleteRequest) (*documentschema.CollectionDeleteResponse, error) {
	db, err := s.getDBFromCtx(ctx, "CollectionDelete")
	if err != nil {
		return nil, err
	}
	resp, err := db.DeleteCollection(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// TODO: implement
func (s *ImmuServer) DocumentAudit(ctx context.Context, req *documentschema.DocumentAuditRequest) (*documentschema.DocumentAuditResponse, error) {
	db, err := s.getDBFromCtx(ctx, "DocumentAudit")
	if err != nil {
		return nil, err
	}
	resp, err := db.DocumentAudit(ctx, req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func (s *ImmuServer) DocumentProof(ctx context.Context, req *documentschema.DocumentProofRequest) (*documentschema.DocumentProofResponse, error) {
	db, err := s.getDBFromCtx(ctx, "DocumentProof")
	if err != nil {
		return nil, err
	}

	res, err := db.DocumentProof(ctx, req)
	if err != nil {
		return nil, err
	}

	if s.StateSigner != nil {
		hdr := schema.TxHeaderFromProto(res.VerifiableTx.DualProof.TargetTxHeader)
		alh := hdr.Alh()

		newState := &schema.ImmutableState{
			Db:     db.GetName(),
			TxId:   hdr.ID,
			TxHash: alh[:],
		}

		err = s.StateSigner.Sign(newState)
		if err != nil {
			return nil, err
		}

		res.VerifiableTx.Signature = newState.Signature
	}

	return res, nil
}

func (s *ImmuServer) CollectionUpdate(ctx context.Context, req *documentschema.CollectionUpdateRequest) (*documentschema.CollectionUpdateResponse, error) {
	db, err := s.getDBFromCtx(ctx, "CollectionUpdate")
	if err != nil {
		return nil, err
	}
	resp, err := db.UpdateCollection(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (s *ImmuServer) DocumentInsertMany(ctx context.Context, req *documentschema.DocumentInsertManyRequest) (*documentschema.DocumentInsertManyResponse, error) {
	db, err := s.getDBFromCtx(ctx, "DocumentInsertMany")
	if err != nil {
		return nil, err
	}
	resp, err := db.DocumentInsertMany(ctx, req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}
