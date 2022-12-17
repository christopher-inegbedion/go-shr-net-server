package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

func InsertUploadedFile(uploadedFile UploadedFile) error {
	if result, err := uploadedFilesColl.InsertOne(context.Background(), uploadedFile); err != nil {
		return err
	} else {
		fmt.Println("Inserted a single document: ", result.InsertedID)
		return nil
	}
}

func GetUploadedFileByFileName(fileName string) (UploadedFile, error) {
	filter := bson.D{{Key: "file_name", Value: fileName}}

	var result UploadedFile
	if err := uploadedFilesColl.FindOne(context.Background(), filter).Decode(&result); err != nil {
		return UploadedFile{}, err
	}
	return result, nil
}

func DeleteUploadedFileByFileName(fileName string) error {
	filter := bson.D{{Key: "file_name", Value: fileName}}
	if _, err := uploadedFilesColl.DeleteOne(context.Background(), filter); err != nil {
		return err
	}
	return nil
}
