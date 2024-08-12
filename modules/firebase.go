package modules

import (
	"context"
	"fmt"
	"io"
	"bytes"
	"time"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go"
	"google.golang.org/api/option"

)

// FirebaseClient is a wrapper around the firebase.App client.
type FirebaseClient struct {
	client *firebase.App
	ctx    context.Context
}

type FS struct {
    client                *storage.Client
    defaultTransferTimeout time.Duration
}

func NewFS(credentialsFilePath string, defaultTransferTimeout time.Duration) (*FS, error) {
    ctx := context.Background()
    client, err := storage.NewClient(ctx, option.WithCredentialsFile(credentialsFilePath))
    if err != nil {
        return nil, err
    }

    return &FS{
        client:                client,
        defaultTransferTimeout: defaultTransferTimeout,
    }, nil
}

// NewFirebaseClient creates a new Firebase client with the provided project ID and secrets JSON.
// It initializes the Firebase app using the given context, project ID, and secrets JSON.
// Returns a pointer to the created FirebaseClient and any error encountered during initialization.
func NewFirebaseClient(ctx context.Context, projectID string, secretsJSON []byte) (*FirebaseClient, error) {
	conf := &firebase.Config{
		ProjectID: projectID,
	}

	opt := option.WithCredentialsJSON(secretsJSON)

	app, err := firebase.NewApp(ctx, conf, opt)
	if err != nil {
		return nil, fmt.Errorf("error initializing app: %v", err)
	}

	return &FirebaseClient{client: app, ctx: ctx}, nil
}

// GetDocument retrieves a document from a specified collection in Firestore.
// It takes the collection name and document ID as parameters and returns the document data as a map[string]interface{} and an error, if any.
func (f *FirebaseClient) GetDocument(collection string, document string) (map[string]interface{}, error) {
	firestoreClient, err := f.client.Firestore(f.ctx)
	if err != nil {
		return nil, err
	}
	defer firestoreClient.Close()

	docRef := firestoreClient.Collection(collection).Doc(document)
	doc, err := docRef.Get(f.ctx)
	if err != nil {
		return nil, err
	}

	return doc.Data(), nil
}

// DeleteDocument deletes a document from the specified collection in Firestore.
// It takes the collection name and document ID as parameters.
// Returns an error if there was a problem deleting the document.
func (f *FirebaseClient) DeleteDocument(collection string, document string) error {
	firestoreClient, err := f.client.Firestore(f.ctx)
	if err != nil {
		return err
	}
	defer firestoreClient.Close()

	docRef := firestoreClient.Collection(collection).Doc(document)
	_, err = docRef.Delete(f.ctx)
	if err != nil {
		return err
	}

	return nil
}

// UpsertDocument inserts or updates a document in the specified collection with the provided data.
// If the document already exists, it will be updated with the new data. If it doesn't exist, a new document will be created.
// The function returns an error if there was a problem with the Firestore operation.
func (f *FirebaseClient) UpsertDocument(collection string, document string, data map[string]interface{}) error {
	firestoreClient, err := f.client.Firestore(f.ctx)
	if err != nil {
		return err
	}
	defer firestoreClient.Close()

	docRef := firestoreClient.Collection(collection).Doc(document)

	_, err = docRef.Set(f.ctx, data, firestore.MergeAll)
	if err != nil {
		return err
	}

	return nil
}

func (f *FirebaseClient) InsertDocument(collection string, data map[string]interface{}) error {
	firestoreClient, err := f.client.Firestore(f.ctx)
	if err != nil {
		return err
	}
	defer firestoreClient.Close()

	_, _, err = firestoreClient.Collection(collection).Add(f.ctx, data)
	if err != nil {
		return err
	}

	return nil
}

func (fs *FS) Upload(fileInput []byte, bucketName, fileName string) error {
    ctx, cancel := context.WithTimeout(context.Background(), fs.defaultTransferTimeout)
    defer cancel()

    bucket := fs.client.Bucket(bucketName)
    object := bucket.Object(fileName)
    writer := object.NewWriter(ctx)
    defer writer.Close()

    if _, err := io.Copy(writer, bytes.NewReader(fileInput)); err != nil {
        return err
    }

    if err := object.ACL().Set(context.Background(), storage.AllUsers, storage.RoleReader); err != nil {
        return err
    }

    return nil
}