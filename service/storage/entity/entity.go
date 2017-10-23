package entity

import (
	"context"
	"fmt"
	"log"
	"time"

	"cloud.google.com/go/datastore"
	"github.com/vjsamuel/uploadly/service/common"
	s "github.com/vjsamuel/uploadly/service/storage"
)

const (
	parent_kind = "Profile"
	entity_kind = "File"
)

type entityStore struct {
	projectId string
	client    *datastore.Client
	ctx       context.Context
}

func NewEntityStorage(projectId string, ctx context.Context) s.Storage {
	client, err := datastore.NewClient(ctx, projectId)
	if err != nil {
		log.Printf("Error instantiating object store client: %v", err)
		return nil
	}

	return &entityStore{client: client, projectId: projectId, ctx: ctx}
}

func (e *entityStore) Get(holder common.Holder) (interface{}, error) {
	parent := e.createAndGetParent(holder)
	if parent == nil {
		return nil, fmt.Errorf("Unable to get parent")
	}
	recordKey := datastore.NameKey(entity_kind, holder.File, parent)

	entity := common.Entity{}
	err := e.client.Get(e.ctx, recordKey, &entity)
	if err != nil {
		log.Printf("Record get failed with error: %v", err)
		return nil, err

	}

	resp := common.Response{
		File: holder.File,
		Size: entity.Size,
		Type: entity.Type,
		UploadTime: entity.UploadTime,
		LastModified: entity.LastModified,
		Version: entity.Version,
		Description: entity.Description,
	}
	return resp, nil
}

func (e *entityStore) Insert(holder common.Holder) error {
	record := common.Entity{
		Size: holder.Size,
		Type: holder.ContentType,
		Version:      1,
		LastModified: time.Now(),
		UploadTime:   time.Now(),
		Description: holder.Description,
	}

	return e.insertRecord(record, holder)
}

func (e *entityStore) Update(holder common.Holder) error {
	rawRecord, err := e.Get(holder)
	if err != nil {
		log.Printf("Unable to find entry to update due to error: %v\n", err)
		return fmt.Errorf("Unable to find entry to update")
	}

	record, _ := rawRecord.(common.Response)
	newRecord := common.Entity{
		Version: record.Version + 1,
		LastModified: time.Now(),
		UploadTime: record.UploadTime,
		Size: holder.Size,
		Type: holder.ContentType,
		Description: holder.Description,
	}
	return e.insertRecord(newRecord, holder)
}

func (e *entityStore) Delete(holder common.Holder) error {
	parent := e.createAndGetParent(holder)
	if parent == nil {
		return fmt.Errorf("Unable to find user profile")
	}
	recordKey := datastore.NameKey(entity_kind, holder.File, parent)

	err := e.client.Delete(e.ctx, recordKey)
	if err != nil {
		log.Printf("Record get delete with error: %v", err)
	}
	return err
}

func (e *entityStore) Exists(holder common.Holder) bool {
	if entity, _ := e.Get(holder); entity != nil {
		return true
	}

	return false
}

func (e *entityStore) List(holder common.Holder) (interface{}, error) {
	parent := e.createAndGetParent(holder)
	if parent == nil {
		return nil, fmt.Errorf("Unable to get parent")
	}

	query := datastore.NewQuery(entity_kind).Ancestor(parent)
	entities := []common.Entity{}
	keys, err := e.client.GetAll(e.ctx, query, &entities)
	if err != nil {
		log.Println("Unable to get list of entries due to error:", err)
		return nil, err
	}

	resp := []common.Response{}
	for i, entity := range entities {
		resp = append(resp, common.Response{
			File: keys[i].Name,
			Type: entity.Type,
			Version: entity.Version,
			LastModified: entity.LastModified,
			UploadTime: entity.UploadTime,
			Size: entity.Size,
			Description: entity.Description,
		})

	}

	return resp, nil
}

func (e *entityStore) createAndGetParent(holder common.Holder) *datastore.Key {
	parent := datastore.NameKey(parent_kind, holder.GetProfileID(), nil)
	profile := common.Profile{}
	if err := e.client.Get(e.ctx, parent, &profile); err != nil {
		// Profile does not exist, create it
		profile = holder.GetProfile()
		_, err := e.client.Put(e.ctx, parent, &profile)
		if err != nil {
			log.Printf("Parent record insert failed with error: %v", err)
			return nil
		}
	}
	return parent
}

func (e *entityStore) insertRecord(record common.Entity, holder common.Holder) error {
	parent := e.createAndGetParent(holder)
	if parent == nil {
		return fmt.Errorf("Unable to find user profile")
	}

	recordKey := datastore.NameKey(entity_kind, holder.File, parent)
	_, err := e.client.Put(e.ctx, recordKey, &record)
	if err != nil {
		log.Printf("Record insert failed with error: %v", err)
		return err
	}

	return nil
}
