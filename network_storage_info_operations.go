package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

var filter = bson.D{{Key: "name", Value: "network-storage-state"}}

// InitialiseNetworkState initialises the network storage state. The network storage stage
// is used to keep track of the total storage size and used storage size of the network, of
// both the AWS storage and the storage pool.
//
// IMPORTANT: This function should only be called once, when the network is first initialised.
func InitialiseNetworkState() {
	storageInfo := NetworkStorageState{
		Name:                     "network-storage-state",
		NumberOfFixedAmount1Subs: 0,
		NumberOfFixedAmount2Subs: 0,
		NumberOfMonthlySubs:      0,
		TotalAwsStorageSize:      0,
		TotalAwsStorageUsed:      0,
		TotalStoragePoolSize:     0,
		TotalStoragePoolUsed:     0,
	}

	result, err := storageCapacityColl.InsertOne(context.TODO(), storageInfo)
	if err != nil {
		panic(err.Error())
	}

	fmt.Printf("Document inserted with ID: %s\n", result.InsertedID)
}

func findNetworkStateInMongo(v *NetworkStorageState) error {
	return storageCapacityColl.FindOne(context.TODO(), filter).Decode(&v)
}

// IsNetworkStorageStateInitialised checks if the network storage state has been initialised.
// If the network storage state has not been initialised, it returns false. If the network
// storage state has been initialised, it returns true.
func IsNetworkStorageStateInitialised() bool {
	var result NetworkStorageState
	err := findNetworkStateInMongo(&result)

	if err == mongo.ErrNoDocuments {
		return false
	}
	if err != nil {
		panic(err)
	}

	return true
}

// GetNetworkStorageState returns the network storage state.
func GetNetworkStorageState() NetworkStorageState {
	var result NetworkStorageState
	err := findNetworkStateInMongo(&result)

	if err == mongo.ErrNoDocuments {
		panic("Network storage state has not been initialised")
	}
	if err != nil {
		panic(err)
	}

	return result
}

// SetTotalAwsStorageSize sets the total AWS storage size in the network.
func IncrementTotalAwsStorageSize(size int64) (bool, error) {
	update := bson.D{
		{Key: "$inc", Value: bson.D{
			{Key: "total_aws_storage_size", Value: size},
		}},
	}

	_, err := storageCapacityColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return false, err
	}

	return true, err
}

// SetTotalStoragePoolSize sets the total storage pool size in the network.
func IncrementTotalStoragePoolSize(size int64) (bool, error) {
	update := bson.D{
		{Key: "$inc", Value: bson.D{
			{Key: "total_storage_pool_size", Value: size},
		}},
	}

	_, err := storageCapacityColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return false, err
	}

	return true, err
}

// IncrementStoragePoolUsed incrementes the storage pool capacity used in the network.
func IncrementStoragePoolUsed(increment int64) (bool, error) {
	update := bson.D{
		{Key: "$inc", Value: bson.D{
			{Key: "total_storage_pool_used", Value: increment},
		}},
	}

	_, err := storageCapacityColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return false, err
	}

	return true, nil
}

// IncrementAwsStorageUsed incrementes the total AWS storage used in the network.
func IncrementAwsStorageUsed(increment int64) (bool, error) {
	update := bson.D{
		{Key: "$inc", Value: bson.D{
			{Key: "total_aws_storage_used", Value: increment},
		}},
	}

	_, err := storageCapacityColl.UpdateOne(context.Background(), filter, update)
	if err != nil {
		return false, err
	}

	return true, nil
}

// GetStoragePoolUsed returns the total storage pool used in the network.
func GetStoragePoolUsed() (int64, error) {
	var result NetworkStorageState
	err := findNetworkStateInMongo(&result)

	if err == mongo.ErrNoDocuments {
		panic("Network storage state has not been initialised")
	}
	if err != nil {
		return 0, err
	}

	return result.TotalStoragePoolUsed, nil
}

// GetTotalAwsStorageUsed returns the total AWS storage used in the network.
func GetTotalAwsStorageUsed() (int64, error) {
	var result NetworkStorageState
	err := findNetworkStateInMongo(&result)

	if err == mongo.ErrNoDocuments {
		panic("Network storage state has not been initialised")
	}
	if err != nil {
		return 0, err
	}

	return result.TotalAwsStorageUsed, nil
}

// GetTotalStoragePoolSize returns the total storage pool size in the network.
func GetTotalStoragePoolSize() (int64, error) {
	var result NetworkStorageState
	err := findNetworkStateInMongo(&result)

	if err == mongo.ErrNoDocuments {
		panic("Network storage state has not been initialised")
	}
	if err != nil {
		return 0, err
	}

	return result.TotalStoragePoolSize, nil
}

// GetTotalAwsStorageSize returns the total AWS storage size in the network.
func GetTotalAwsStorageSize() int64 {
	var result NetworkStorageState
	err := findNetworkStateInMongo(&result)

	if err == mongo.ErrNoDocuments {
		panic("Network storage state has not been initialised")
	}
	if err != nil {
		panic(err)
	}

	return result.TotalAwsStorageSize
}

// GetNumberOfMonthlySubs returns the number of customers on the monthly subscription plan.
func GetNumberOfMonthlySubs() int64 {
	var result NetworkStorageState
	err := findNetworkStateInMongo(&result)

	if err == mongo.ErrNoDocuments {
		panic("Network storage state has not been initialised")
	}
	if err != nil {
		panic(err)
	}

	return result.NumberOfMonthlySubs
}

// GetNumberOfFixedAmount1Subs returns the number of customers on the fixed amount 1TB subscription plan.
func GetNumberOfFixedAmount1Subs() (int64, error) {
	var result NetworkStorageState
	err := findNetworkStateInMongo(&result)

	if err == mongo.ErrNoDocuments {
		panic("Network storage state has not been initialised")
	}
	if err != nil {
		return 0, err
	}

	return result.NumberOfFixedAmount1Subs, nil
}

// GetNumberOfFixedAmount2Subs returns the number of customers on the fixed amount 2TB subscription plan.
func GetNumberOfFixedAmount2Subs() (int64, error) {
	var result NetworkStorageState
	err := findNetworkStateInMongo(&result)

	if err == mongo.ErrNoDocuments {
		panic("Network storage state has not been initialised")
	}
	if err != nil {
		return 0, err
	}

	return result.NumberOfFixedAmount2Subs, nil
}
