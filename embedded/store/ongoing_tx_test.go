/*
Copyright 2022 Codenotary Inc. All rights reserved.

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

package store

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestOngoingTXAddPrecondition(t *testing.T) {
	otx := OngoingTx{
		st: &ImmuStore{
			maxKeyLen: 10,
		},
	}

	err := otx.AddPrecondition(nil)
	require.ErrorIs(t, err, ErrIllegalArguments)

	err = otx.AddPrecondition(&PreconditionKeyMustExist{})
	require.ErrorIs(t, err, ErrInvalidPrecondition)

	otx.closed = true
	err = otx.AddPrecondition(&PreconditionKeyMustExist{
		Key: []byte("key"),
	})
	require.ErrorIs(t, err, ErrAlreadyClosed)
}

func TestOngoingTxCheckPreconditionsCornerCases(t *testing.T) {
	st, err := Open(t.TempDir(), DefaultOptions())
	require.NoError(t, err)

	defer immustoreClose(t, st)

	otx := &OngoingTx{}

	err = otx.checkPreconditions(st)
	require.NoError(t, err)

	otx.preconditions = []Precondition{nil}
	err = otx.checkPreconditions(st)
	require.ErrorIs(t, err, ErrInvalidPrecondition)
	require.ErrorIs(t, err, ErrInvalidPreconditionNull)

	err = st.Close()
	require.NoError(t, err)

	otx.preconditions = []Precondition{
		&PreconditionKeyMustExist{Key: []byte{1}},
	}
	err = otx.checkPreconditions(st)
	require.ErrorIs(t, err, ErrAlreadyClosed)

	otx.preconditions = []Precondition{
		&PreconditionKeyMustNotExist{Key: []byte{1}},
	}
	err = otx.checkPreconditions(st)
	require.ErrorIs(t, err, ErrAlreadyClosed)

	otx.preconditions = []Precondition{
		&PreconditionKeyNotModifiedAfterTx{Key: []byte{1}, TxID: 1},
	}
	err = otx.checkPreconditions(st)
	require.ErrorIs(t, err, ErrAlreadyClosed)
}

func TestOngoingTxOptions(t *testing.T) {
	var opts *TxOptions
	require.Error(t, opts.Validate())

	opts = &TxOptions{}
	require.Equal(t, TxMode(4), opts.WithMode(4).Mode)
	require.Error(t, opts.Validate())

	require.Equal(t, 1*time.Hour, opts.WithSnapshotRenewalPeriod(1*time.Hour).SnapshotRenewalPeriod)
}
