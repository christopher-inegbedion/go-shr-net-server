package main

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// InsertUser inserts the given user into the database.
func InsertUser(user User) (bool, error) {
	// Check if the user already exists in the database.
	if _, err := GetUserByAddress(user.Address); err == nil {
		return false, fmt.Errorf("user already exists")
	}

	if _, err := userDetailsColl.InsertOne(context.Background(), user); err != nil {
		return false, err
	} else {
		if user.AccountType == MONTHLY_SUB {
			if ok, err := IncrementTotalStoragePoolSize(MONTHLY_STORAGE_SIZE); err != nil {
				return false, err
			} else if !ok {
				return false, fmt.Errorf("failed to increment storage pool")
			} else {
				return true, nil
			}
		} else if user.AccountType == FIXED_AMOUNT_1 {
			if ok, err := IncrementTotalAwsStorageSize(FIXED_AMOUNT_1_STORAGE_SIZE); err != nil {
				return false, err
			} else if !ok {
				return false, fmt.Errorf("failed to increment aws storage")
			} else {
				return true, nil
			}
		} else if user.AccountType == FIXED_AMOUNT_2 {
			if ok, err := IncrementTotalAwsStorageSize(FIXED_AMOUNT_2_STORAGE_SIZE); err != nil {
				return false, err
			} else if !ok {
				return false, fmt.Errorf("failed to increment aws storage")
			} else {
				return true, nil
			}
		} else {
			return false, fmt.Errorf("invalid account type")
		}
	}
}

// UpdateUser updates the user with the given address.
func UpdateUser(fieldName string, fieldValue interface{}, address string) (bool, error) {
	// Check if the user exists in the database.
	if _, err := GetUserByAddress(address); err != nil {
		return false, err
	} 

	filter := bson.D{{Key: "address", Value: address}}

	var update bson.D

	if fieldName == "spool_capacity_used" || fieldName == "aws_capacity_used" || fieldName == "number_files_uploaded" {
		update = bson.D{{Key: "$inc", Value: bson.D{{Key: fieldName, Value: fieldValue}}}}

		if fieldName == "spool_capacity_used" {
			if ok, err := IncrementStoragePoolUsed(fieldValue.(float64)); err != nil {
				return false, err
			} else if !ok {
				return false, fmt.Errorf("failed to increment storage pool")
			}
		} else if fieldName == "aws_capacity_used" {
			if ok, err := IncrementAwsStorageUsed(fieldValue.(float64)); err != nil {
				return false, err
			} else if !ok {
				return false, fmt.Errorf("failed to increment aws storage")
			}
		}
	} else {
		update = bson.D{{Key: "$set", Value: bson.D{{Key: fieldName, Value: fieldValue}}}}
	}

	if _, err := userDetailsColl.UpdateOne(context.Background(), filter, update); err != nil {
		return false, err
	} else {
		return true, nil
	}
}

// DeleteUser deletes the user with the given address.
func DeleteUser(address string) (bool, error) {
	filter := bson.D{{Key: "address", Value: address}}

	if user, err := GetUserByAddress(address); err != nil {
		return false, err
	} else {
		if _, err := userDetailsColl.DeleteOne(context.Background(), filter); err != nil {
			return false, err
		} else {
			if user.AccountType == MONTHLY_SUB {
				if ok, err := DecrementTotalStoragePoolSize(MONTHLY_STORAGE_SIZE); err != nil {
					return false, err
				} else if !ok {
					return false, fmt.Errorf("failed to increment storage pool")
				} else {
					// Decrement the storage pool used.
					if ok, err := DecrementStoragePoolUsed(user.SpoolCapacityUsed); err != nil {
						return false, err
					} else if !ok {
						return false, fmt.Errorf("failed to decrement storage pool")
					} else {
						return true, nil
					}
				}
			} else if user.AccountType == FIXED_AMOUNT_1 {
				if ok, err := DecrementTotalAwsStorageSize(FIXED_AMOUNT_1_STORAGE_SIZE); err != nil {
					return false, err
				} else if !ok {
					return false, fmt.Errorf("failed to increment aws storage")
				} else {
					return true, nil
				}
			} else if user.AccountType == FIXED_AMOUNT_2 {
				if ok, err := DecrementTotalAwsStorageSize(FIXED_AMOUNT_2_STORAGE_SIZE); err != nil {
					return false, err
				} else if !ok {
					return false, fmt.Errorf("failed to increment aws storage")
				} else {
					// Decrement the storage pool used.
					if ok, err := DecrementAwsStorageUsed(user.SpoolCapacityUsed); err != nil {
						return false, err
					} else if !ok {
						return false, fmt.Errorf("failed to decrement storage pool")
					} else {
						return true, nil
					}
				}
			} else {
				return false, fmt.Errorf("invalid account type")
			}
		}
	}
}

// GetUserByAddress returns the user with the given address.
func GetUserByAddress(address string) (User, error) {
	filter := bson.D{{Key: "address", Value: address}}

	var result User
	if err := userDetailsColl.FindOne(context.Background(), filter).Decode(&result); err != nil {
		if err == mongo.ErrNoDocuments {
			return User{}, fmt.Errorf("user not found")
		}
	}

	return result, nil
}
