package memstore_test

import (
	"context"
	"kvstore/memstore"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type keyValue struct {
	key     string
	value   string
	wantErr bool
}

func TestCreate(t *testing.T) {
	type testCaseCreate struct {
		name string
		kv   []*keyValue
	}
	testCases := []testCaseCreate{
		{
			name: "ok",
			kv: []*keyValue{
				{
					key:   "k",
					value: "v",
				},
			},
		},
		{
			name: "empty key",
			kv: []*keyValue{
				{
					key:     "",
					value:   "v",
					wantErr: true,
				},
			},
		},
		{
			name: "key already exists",
			kv: []*keyValue{
				{
					key:   "k",
					value: "v",
				},
				{
					key:     "k",
					value:   "v",
					wantErr: true,
				},
			},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := memstore.New()
			for _, kv := range tc.kv {
				err := store.Create(ctx, []byte(kv.key), []byte(kv.value))
				if kv.wantErr && err == nil {
					t.Errorf("Create(%v, %v, %v): returned err=nil, want not nil", ctx, kv.key, kv.value)
				}
				if !kv.wantErr && err != nil {
					t.Errorf("Create(%v, %v, %v): returned err=%v, want nil", ctx, kv.key, kv.value, err)
				}
			}
		})
	}
}

func TestRead(t *testing.T) {
	type testCaseRead struct {
		name    string
		create  *keyValue
		key     string
		want    []byte
		wantErr bool
	}
	testCases := []testCaseRead{
		{
			name: "ok",
			create: &keyValue{
				key:   "k",
				value: "v",
			},
			key:  "k",
			want: []byte("v"),
		},
		{
			name:    "key does not exist",
			key:     "k",
			wantErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := memstore.New()
			if tc.create != nil {
				if err := store.Create(ctx, []byte(tc.create.key), []byte(tc.create.value)); err != nil {
					t.Fatalf("Create(%v, %v, %v): returned err=%v, want nil", ctx, tc.create.key, tc.create.value, err)
				}
			}
			got, err := store.Read(ctx, []byte(tc.key))
			if !tc.wantErr && err != nil {
				t.Errorf("Read(%v, %v): returned err=%v, want nil", ctx, tc.key, err)
			}
			if tc.wantErr && err == nil {
				t.Errorf("Read(%v, %v): returned err=nil, want not nil", ctx, tc.key)
			}
			if diff := cmp.Diff(got, tc.want); diff != "" {
				t.Errorf("Read(%v, %v): returned unexpected diff (-got +want):\n%s", ctx, tc.key, diff)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	type testCaseUpdate struct {
		name          string
		create        string
		createKey     string
		update        string
		updateKey     string
		wantUpdateErr bool
		want          string
	}
	testCases := []testCaseUpdate{
		{
			name:      "ok",
			create:    "v",
			createKey: "k",
			update:    "v1",
			updateKey: "k",
			want:      "v1",
		},
		{
			name:          "key does not exist",
			create:        "v",
			update:        "v1",
			createKey:     "k",
			updateKey:     "k1",
			wantUpdateErr: true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := memstore.New()
			if err := store.Create(ctx, []byte(tc.createKey), []byte(tc.create)); err != nil {
				t.Fatal(err)
			}
			err := store.Update(ctx, []byte(tc.updateKey), []byte(tc.update))
			if !tc.wantUpdateErr && err != nil {
				t.Errorf("Update(%v, %v, %v) returned err=%v, want nil", ctx, tc.updateKey, tc.update, err)
			}
			if tc.wantUpdateErr {
				if err == nil {
					t.Errorf("Update(%v, %v, %v) returned err=nil, want not nil", ctx, tc.updateKey, tc.update)
				}
				return
			}
			got, err := store.Read(ctx, []byte(tc.updateKey))
			if err != nil {
				t.Errorf("Read(%v, %v) returned err=%v, want nil", ctx, tc.updateKey, err)
			}
			want := []byte(tc.want)
			if diff := cmp.Diff(got, want, cmpopts.EquateEmpty()); diff != "" {
				t.Errorf("Read(%v, %v) returned unexpected diff (-got +want):\n%s", ctx, tc.updateKey, diff)
			}
		})
	}
}

func TestDelete(t *testing.T) {
	type testCaseDelete struct {
		name        string
		createKey   string
		createValue string
		deleteKey   string
		wantErr     bool
	}
	testCases := []testCaseDelete{
		{
			name:        "ok",
			createKey:   "k",
			createValue: "v",
			deleteKey:   "k",
		},
		{
			name:        "key not found",
			createKey:   "k",
			createValue: "v",
			deleteKey:   "k1",
			wantErr:     true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx := context.Background()
			store := memstore.New()
			if err := store.Create(ctx, []byte(tc.createKey), []byte(tc.createValue)); err != nil {
				t.Fatal(err)
			}
			err := store.Delete(ctx, []byte(tc.deleteKey))
			if !tc.wantErr && err != nil {
				t.Errorf("Delete(%v, %v) returned err=%v, want nil", ctx, tc.deleteKey, err)
			}
			if tc.wantErr && err == nil {
				t.Errorf("Delete(%v, %v) returned err=nil, want not nil", ctx, tc.deleteKey)
			}
			if got, err := store.Read(ctx, []byte(tc.deleteKey)); err == nil {
				t.Errorf("Read(%v, %v) returned %v, want nil", ctx, tc.deleteKey, got)
			}
		})
	}
}

func FuzzCRUD(f *testing.F) {
	ctx := context.Background()
	f.Add([]byte("key"), []byte("value1"), []byte("value2"))
	f.Fuzz(func(t *testing.T, key, value1, value2 []byte) {
		store := memstore.New()
		if err := store.Create(ctx, key, value1); err != nil {
			return
		}
		got, err := store.Read(ctx, key)
		if err != nil {
			t.Errorf("Read(%v, %v) returned err=%v, want nil", ctx, key, err)
		}
		want := value1
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("Read(%v, %v) returned unexpected diff (-got +want):\n%s", ctx, key, diff)
		}
		if err := store.Update(ctx, key, value2); err != nil {
			t.Errorf("Update(%v, %v, %v) returned err=%v, want nil", ctx, key, value2, err)
		}
		got, err = store.Read(ctx, key)
		if err != nil {
			t.Errorf("Read(%v, %v) returned err=%v, want nil", ctx, key, err)
		}
		want = value2
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("Read(%v, %v) returned unexpected diff (-got +want):\n%s", ctx, key, diff)
		}
	})
}
