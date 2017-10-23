package object

import (
	"context"
	"fmt"
	"log"
	"io"

	"cloud.google.com/go/storage"
	s "github.com/vjsamuel/uploadly/service/storage"
	"github.com/vjsamuel/uploadly/service/common"
	"google.golang.org/api/iterator"
)

type objectStore struct {
	bucket    string
	projectId string
	client    *storage.Client
	ctx       context.Context
}

func NewObjectStorage(bucket string, projectId string, ctx context.Context) s.Storage {
	client, err := storage.NewClient(ctx)
	if err != nil {
		log.Printf("Error instantiating object store client: %v", err)
		return nil
	}

	buck := client.Bucket(bucket)
	if _, err := buck.Attrs(ctx); err != nil {
		e := buck.Create(ctx, projectId, nil)
		if e != nil {
			log.Fatal("Unable to create bucket due to error: ", err)
		}
	}
	return &objectStore{bucket: bucket, client: client, projectId: projectId, ctx: ctx}
}

func (o *objectStore) Get(holder common.Holder) (interface{}, error) {
	if o.Exists(holder) == false {
		return nil, nil
	}
	id := holder.GetProfileID()
	buck := o.client.Bucket(o.bucket)
	obj := buck.Object(fmt.Sprintf("%s/%s", id, holder.File))
	reader, err := obj.NewReader(o.ctx)
	if err != nil {
		return nil, err
	}

	return reader, nil
}

func (o *objectStore) Insert(holder common.Holder) error {
	id := holder.GetProfileID()
	buck := o.client.Bucket(o.bucket)
	obj := buck.Object(fmt.Sprintf("%s/%s", id, holder.File))

	writer := obj.NewWriter(o.ctx)
	reader, ok := holder.Object.(io.Reader)
	if !ok {
		return fmt.Errorf("Unable to get Reader for input object")
	}
	_, err := io.Copy(writer, reader)
	if err != nil {
		return fmt.Errorf("Unable to write object due to error: %v", err)
	}
	writer.Close()

	return nil
}

func (o *objectStore) Update(holder common.Holder) error {
	return o.Insert(holder)
}

func (o *objectStore) Delete(holder common.Holder) error {
	if o.Exists(holder) == false {
		return nil
	}
	id := holder.GetProfileID()
	buck := o.client.Bucket(o.bucket)
	obj := buck.Object(fmt.Sprintf("%s/%s", id, holder.File))

	err := obj.Delete(o.ctx)
	return err
}

func (o *objectStore) Exists(holder common.Holder) bool {
	id := holder.GetProfileID()
	buck := o.client.Bucket(o.bucket)
	obj := buck.Object(fmt.Sprintf("%s/%s", id, holder.File))

	if attr, _ := obj.Attrs(o.ctx); attr != nil {
		return true
	}
	return false
}

func (o *objectStore) List(holder common.Holder) (interface{}, error) {
	id := holder.GetProfileID()
	buck := o.client.Bucket(o.bucket)

	it := buck.Objects(o.ctx, &storage.Query{
		Prefix: id,
	})

	resps := []common.Response{}
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		resps = append(resps, common.Response{
			Version: 1,
			LastModified: attrs.Updated,
			UploadTime: attrs.Created,
			Size: attrs.Size,
			File: attrs.Name,
			Type: attrs.ContentType,
		})
	}

	return resps, nil
}
