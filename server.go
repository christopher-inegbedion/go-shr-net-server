package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var ServentWSAdddress = fmt.Sprintf("0.0.0.0:%v", PORT)
var client *mongo.Client
var storageCapacityColl *mongo.Collection
var uploadedFilesColl *mongo.Collection
var userDetailsColl *mongo.Collection

var RouteCommands map[string]func(http.ResponseWriter, *http.Request) = make(map[string]func(http.ResponseWriter, *http.Request))

// CreateCommandAction creates an action to be executed when a path is requested
func CreateCommandAction(path string, action func(http.ResponseWriter, *http.Request)) {
	if path != "" && action != nil {
		RouteCommands[path] = action
	}
}

// StartServer starts the server
func StartServer() {
	if c, err := InstantiateMongoDB(); err != nil {
		panic(err)
	} else {
		client = c
		storageCapacityColl = client.Database(DB_NAME).Collection(STORAGE_CAPACITY_COLL_NAME)
		uploadedFilesColl = client.Database(DB_NAME).Collection(UPLOADED_FILES_COLL_NAME)
		userDetailsColl = client.Database(DB_NAME).Collection(USER_DETAILS_COLL_NAME)

		defer func() {
			if err := client.Disconnect(context.TODO()); err != nil {
				panic(err)
			}
		}()
	}

	println("Server started on port", PORT)

	CreateCommandAction("/init", initialiseStorageStateHandler)

	// Routes for getting the total AWS and storage pool size
	CreateCommandAction("/size/aws", getAwsStorageSizeHandler)
	CreateCommandAction("/size/spool", getTotalStoragePoolSizeHandler)

	// Routes for getting the total AWS and storage pool used (GET)
	// and incrementing the total AWS and storage pool used (POST request)
	CreateCommandAction("/used/aws", getAwsStorageUsedHandler)
	CreateCommandAction("/used/spool", getStoragePoolUsedHandler)

	// Route for instructing the node how to store the file
	CreateCommandAction("/store", storeFileHandler)

	// Route to record the uploaded file
	CreateCommandAction("/file", recordFileHandler)

	// Route to increment the total AWS and storage pool size
	CreateCommandAction("/inc/aws", incrementAwsStorageSizeHandler)
	CreateCommandAction("/inc/spool", incrementStoragePoolSizeHandler)

	// Route to manage the users
	CreateCommandAction("/user", manageUserHandler)

	CreateCommandAction("/users", getUsersHandler)

	for path, action := range RouteCommands {
		http.HandleFunc(path, action)
	}

	// Listens for incoming connections and runs their handler
	if err := http.ListenAndServe(ServentWSAdddress, nil); err != nil {
		log.Fatal(err.Error())
	}
}

func initialiseStorageStateHandler(w http.ResponseWriter, r *http.Request) {
	InitialiseNetworkState()
}

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		queryParams := r.URL.Query()
		amount := queryParams.Get("amount")
		if amount == "" {
			SendResponse(w, false, "amount form key not provided", nil)
			return
		}

		amountInt, _ := strconv.Atoi(amount)

		if users, err := GetUsers(amountInt); err != nil {
			SendResponse(w, false, err.Error(), nil)
		} else {
			SendResponse(w, true, "Users", users)
		}
	}
}

// manageUserHandler manages the users
func manageUserHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
		log.Println(r.Method)

	switch r.Method {
	case "GET":
		queryParams := r.URL.Query()
		address := queryParams.Get("address")
		if address == "" {
			SendResponse(w, false, "address form key not provided", nil)
			return
		}

		if user, err := GetUserByAddress(address); err != nil {
			SendResponse(w, false, err.Error(), nil)
		} else {
			SendResponse(w, true, "User details", user)
		}

	case "PUT":
		address := r.FormValue("address")
		fieldName := r.FormValue("field_name")
		fieldValue := r.FormValue("field_value")

		if address == "" || fieldName == "" || fieldValue == "" {
			SendResponse(w, false, "Invalid parameters", nil)
			return
		}

		var value interface{}
		value = fieldValue

		if fieldName == "spool_capacity_used" || fieldName == "aws_capacity_used" {
			value, _ = strconv.Atoi(fieldValue)
			value = int64(value.(int))
		}

		if fieldName == "num_files_uploaded" {
			value, _ = strconv.Atoi(fieldValue)
			value = int(value.(int))
		}

		if ok, err := UpdateUser(fieldName, value, address); err != nil {
			SendResponse(w, false, err.Error(), nil)
		} else {
			if ok {
				SendResponse(w, true, "User field updated", nil)
			} else {
				SendResponse(w, false, "User field failed", nil)
			}
		}

	case "POST":
		log.Println("POST request received")
		address := r.FormValue("address")
		userName := r.FormValue("user_name")
		timezone := r.FormValue("timezone")
		accountType := r.FormValue("account_type")
		spoolCapacityUsed, _ := strconv.Atoi(r.FormValue("spool_capacity_used"))
		awsCapacityUsed, _ := strconv.Atoi(r.FormValue("aws_capacity_used"))
		numFilesUploaded, _ := strconv.Atoi(r.FormValue("num_files_uploaded"))

		if address == "" || userName == "" || timezone == "" || accountType == "" {
			SendResponse(w, false, "Invalid parameters", nil)
			return
		}

		if accountType != MONTHLY_SUB && accountType != FIXED_AMOUNT_1 && accountType != FIXED_AMOUNT_2 {
			SendResponse(w, false, fmt.Sprintf("Invalid account type [%v]", accountType), nil)
			return
		}

		user := User{
			Address:           address,
			UserName:          userName,
			Timezone:          timezone,
			AccountType:       accountType,
			SpoolCapacityUsed: float64(spoolCapacityUsed),
			AwsCapacityUsed:   float64(awsCapacityUsed),
			NumFilesUploaded:  numFilesUploaded,
		}

		if ok, err := InsertUser(user); err != nil {
			SendResponse(w, false, err.Error(), nil)
		} else {
			if ok {
				SendResponse(w, true, "User added", nil)
			} else {
				SendResponse(w, false, "User not inserted", nil)
			}
		}
	case "DELETE":
		if r.FormValue("address") == "" {
			SendResponse(w, false, "address form key not provided", nil)
			return
		}

		if ok, err := DeleteUser(r.FormValue("address")); err != nil {
			SendResponse(w, false, err.Error(), nil)
		} else {
			if ok {
				SendResponse(w, true, "User deleted", nil)
			} else {
				SendResponse(w, false, "User not deleted", nil)
			}
		}
	}
}

// getTotalStoragePoolSizeHandler returns the total storage pool size
func getTotalStoragePoolSizeHandler(w http.ResponseWriter, r *http.Request) {
	if totalStoragePoolSize, err := GetTotalStoragePoolSize(); err != nil {
		SendResponse(w, false, err.Error(), nil)
	} else {
		SendResponse(w, true, "Total storage pool size", totalStoragePoolSize)
	}
}

// incrementAwsStorageSizeHandler increments the total AWS storage size
func incrementAwsStorageSizeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if r.FormValue("amount") == "" {
		SendResponse(w, false, "amount form key not provided", nil)
		return
	}

	newAwsStorage, _ := strconv.Atoi(r.FormValue("amount"))

	if ok, err := IncrementTotalAwsStorageSize(float64(newAwsStorage)); err != nil {
		SendResponse(w, false, err.Error(), nil)
	} else {
		if ok {
			SendResponse(w, true, "AWS storage incremented", nil)
		} else {
			SendResponse(w, false, "AWS storage not incremented", nil)
		}
	}
}

func incrementStoragePoolSizeHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if r.FormValue("amount") == "" {
		SendResponse(w, false, "amount form key not provided", nil)
		return
	}

	newStoragePool, _ := strconv.Atoi(r.FormValue("amount"))

	if ok, err := IncrementTotalStoragePoolSize(float64(newStoragePool)); err != nil {
		SendResponse(w, false, err.Error(), nil)
	} else {
		if ok {
			SendResponse(w, true, "Storage pool incremented", nil)
		} else {
			SendResponse(w, false, "Storage pool not incremented", nil)
		}
	}
}

// recordFileHandler is called after a file has been uploaded. It records the file
// in the database.
func recordFileHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "POST":
		r.ParseForm()

		fileName := r.FormValue("file_name")
		fileSize, _ := strconv.ParseFloat(r.FormValue("file_size"), 64)
		uploadDate, _ := strconv.Atoi(r.FormValue("upload_date"))
		inStoragePool := r.FormValue("in_storage_pool")
		hosts := r.FormValue("hosts")
		shards := r.FormValue("shards")
		uploaderAddress := r.FormValue("uploader_address")
		backupShards := r.FormValue("backup_shards")
		isMonthlySub := r.FormValue("is_monthly_sub")
		timezone := r.FormValue("timezone")

		if fileName == "" || uploadDate == 0 || inStoragePool == "" || hosts == "" || uploaderAddress == "" || isMonthlySub == "" || timezone == "" {
			SendResponse(w, false, "Invalid parameters", nil)
			return
		}

		// Check if the user exists
		if _, err := GetUserByAddress(uploaderAddress); err != nil {
			SendResponse(w, false, err.Error(), nil)
			panic(err)
			return
		}

		// Convert the hosts string to a 2D string array
		var hosts2D [][]string
		if err := json.Unmarshal([]byte(hosts), &hosts2D); err != nil {
			SendResponse(w, false, err.Error(), nil)
			panic(err)
			return
		}

		// Convert the inStoragePool string to a boolean
		var inStoragePoolBool bool
		if inStoragePool == "true" {
			inStoragePoolBool = true
		} else {
			inStoragePoolBool = false
		}

		// Convert the shards string to an int
		var shardsInt int
		if shards == "" {
			shardsInt = 0
		} else {
			shardsInt, _ = strconv.Atoi(shards)
		}

		// Convert the backupShards string to an int
		var backupShardsInt int
		if backupShards == "" {
			backupShardsInt = 0
		} else {
			backupShardsInt, _ = strconv.Atoi(backupShards)
		}

		// Convert the isMonthlySub string to a boolean
		var isMonthlySubBool bool
		if isMonthlySub == "true" {
			isMonthlySubBool = true
		} else {
			isMonthlySubBool = false
		}

		uploadedFile := UploadedFile{
			FileName:        fileName,
			FileSize:        float64(fileSize),
			UploadDate:      uploadDate,
			InStoragePool:   inStoragePoolBool,
			Hosts:           hosts2D,
			Shards:          shardsInt,
			UploaderAddress: uploaderAddress,
			BackupShards:    backupShardsInt,
			IsMonthlySub:    isMonthlySubBool,
			Timezone:        timezone,
		}

		// Increment the storage pool used
		if err := InsertUploadedFile(uploadedFile); err != nil {
			SendResponse(w, false, err.Error(), nil)
		} else {
			var fieldToUpdate string
			var successMsg string
			var failMsg string
			if inStoragePoolBool {
				fieldToUpdate = "spool_capacity_used"
				successMsg = "File upload success (Storage Pool)"
				failMsg = "File upload success (Storage Pool), error updating user"
			} else {
				fieldToUpdate = "aws_capacity_used"
				successMsg = "File upload success (AWS)"
				failMsg = "File upload success (AWS), error updating user"
			}

			if ok, err := UpdateUser(fieldToUpdate, float64(fileSize), uploaderAddress); err != nil {
				SendResponse(w, false, err.Error(), nil)
			} else {
				if ok {
					SendResponse(w, true, successMsg, nil)
				} else {
					SendResponse(w, false, failMsg, nil)
				}
			}
		}
	}

}

// storeFileHandler is called before a file is to be uploaded. It tells the node
// where to store the file and if they can store it.
//
// If the user is a monthly subscriber, then they should be told to store their file
// in AWS except in situations where the storage pool can satisfy all the Fixed Amount
// customers, in that case their file would be stored in the storage pool.
// If the user is a Fixed Amount customer, then they should be told to store their file
// in the storage pool except in situations where the storage pool is full, in that case
// their file would be stored in AWS.
func storeFileHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()

	if r.FormValue("file_size_gb") == "" {
		SendResponse(w, false, "file_size_gb form key not specified", nil)
		return
	}

	if r.FormValue("account_type") == "" {
		SendResponse(w, false, "account_type form key not specified", nil)
		return
	}

	fileSizeGB, _ := strconv.Atoi(r.FormValue("file_size_gb"))
	accountType := r.FormValue("account_type")

	totalStoragePoolSize, err := GetTotalStoragePoolSize()
	if err != nil {
		SendResponse(w, false, err.Error(), nil)
	}

	if accountType == "monthly" {
		// Check if the storage pool can satisfy the file
		numberOfFixedAmounts1, err := GetNumberOfFixedAmount1Subs()
		if err != nil {
			SendResponse(w, false, err.Error(), nil)
		}

		numberOfFixedAmounts2, err := GetNumberOfFixedAmount2Subs()
		if err != nil {
			SendResponse(w, false, err.Error(), nil)
		}

		if float64(numberOfFixedAmounts1*1000) >= totalStoragePoolSize+float64(fileSizeGB) {
			// Storage pool cannot satisfy the file
			SendResponse(w, true, "location", "aws")
		} else if float64(numberOfFixedAmounts2*2000) >= totalStoragePoolSize+float64(fileSizeGB) {
			// Storage pool cannot satisfy the file
			SendResponse(w, true, "location", "aws")
		} else {
			// Storage pool can satisfy the file
			SendResponse(w, true, "location", "spool")
		}
	} else if accountType == "fixed1" {
		// Check if the storage pool can satisfy the file
		if float64(fileSizeGB) < totalStoragePoolSize {
			// Storage pool can satisfy the file
			SendResponse(w, true, "location", "spool")
		} else {
			// Storage pool cannot satisfy the file
			SendResponse(w, true, "location", "aws")
		}
	} else if accountType == "fixed2" {
		// Check if the storage pool can satisfy the file
		if float64(fileSizeGB) < totalStoragePoolSize {
			// Storage pool can satisfy the file
			SendResponse(w, true, "location", "spool")
		} else {
			// Storage pool cannot satisfy the file
			SendResponse(w, true, "location", "aws")
		}
	} else {
		SendResponse(w, false, fmt.Sprintf("Invalid account type [%v]", accountType), nil)
	}
}

// getStoragePoolUsedHandler handles sending the total storage pool used when a GET request is made
// and handles incrementing the total storage pool used when a POST request is made
func getStoragePoolUsedHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET": // Get storage pool used
		storagePoolUsed, err := GetStoragePoolUsed()
		if err != nil {
			SendResponse(w, false, err.Error(), nil)
		} else {
			SendResponse(w, true, "Total storage pool used", storagePoolUsed)
		}

	case "POST": // Increment storage pool used
		r.ParseForm()
		size, _ := strconv.Atoi(r.FormValue("size"))

		// Ensure that the storage pool is not full
		if ok, err := updateStoragePoolUsed(float64(size)); err != nil {
			SendResponse(w, false, err.Error(), nil)
		} else {
			if ok {
				SendResponse(w, true, "Storage pool used incremented", nil)
			} else {
				SendResponse(w, false, "Storage pool is full", nil)
			}
		}
	}
}

func updateStoragePoolUsed(fileSize float64) (bool, error) {
	// Update the storage pool used
	storagePoolAvailable, err := GetTotalStoragePoolSize()
	if err != nil {
		return false, err
	}

	totalStoragePoolUsed, err := GetStoragePoolUsed()
	if err != nil {
		return false, err
	}

	if storagePoolAvailable-totalStoragePoolUsed <= float64(fileSize) {
		fmt.Println("Storage pool is full")
		return false, nil
	}

	// Increment the storage pool used
	if ok, err := IncrementStoragePoolUsed(float64(fileSize)); err != nil {
		return false, err
	} else {
		if !ok {
			return false, fmt.Errorf("failed to increment storage pool used")
		}
	}

	return true, nil
}

func updateAWSstorageUsed(fileSize float64) (bool, error) {
	// Update the AWS storage used
	awsStorageAvailable, err := GetTotalAwsStorageSize()
	if err != nil {
		return false, err
	}

	totalAwsStorageUsed, err := GetTotalAwsStorageUsed()
	if err != nil {
		return false, err
	}

	if awsStorageAvailable-totalAwsStorageUsed <= float64(fileSize) {
		fmt.Println("AWS storage is full")
		return false, nil
	}

	// Increment the AWS storage used
	if ok, err := IncrementAwsStorageUsed(float64(fileSize)); err != nil {
		return false, err
	} else {
		if !ok {
			return false, fmt.Errorf("failed to increment AWS storage used")
		}
	}

	return true, nil
}

func getAwsStorageSizeHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		if size, err := GetTotalAwsStorageSize(); err == nil {
			SendResponse(w, true, "Total AWS storage size", size)
		} else {
			SendResponse(w, false, err.Error(), nil)
		}

	case "POST":
		r.ParseForm()
		size, _ := strconv.Atoi(r.FormValue("size"))

		if success, err := IncrementAwsStorageUsed(float64(size)); err != nil {
			SendResponse(w, false, err.Error(), nil)
		} else {
			totalAwsUsed, err := GetTotalAwsStorageUsed()
			if err != nil {
				SendResponse(w, false, err.Error(), nil)
			}

			SendResponse(w, success, "Total AWS storage used", totalAwsUsed)
		}

	default:
		fmt.Println("Invalid request method")
	}
}

func getAwsStorageUsedHandler(w http.ResponseWriter, r *http.Request) {
	totalAwsStorageUsed, err := GetTotalAwsStorageUsed()
	if err != nil {
		SendResponse(w, false, err.Error(), nil)
	}

	SendResponse(w, true, "Total AWS storage used", totalAwsStorageUsed)
}

// SendResponse sends a response to the requester
func SendResponse(w http.ResponseWriter, success bool, message string, value interface{}) {
	response := Response{
		Message: message,
		Data:    value,
		Success: success,
	}

	jsonData, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		panic(err)
	}

	fmt.Fprint(w, string(jsonData))
}

func InstantiateMongoDB() (*mongo.Client, error) {
	uri := "mongodb+srv://admin:eromosele1234@shr.nsdaozg.mongodb.net/?retryWrites=true&w=majority"

	client, err := mongo.Connect(context.TODO(), options.Client().ApplyURI(uri))
	if err != nil {
		panic(err)
	}

	return client, err
}
