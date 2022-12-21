package main

const (
	PORT = 12345
)

// MongoDB constants
const (
	DB_NAME = "shr-network-information"

	STORAGE_CAPACITY_COLL_NAME = "storage-capacity-info"
	UPLOADED_FILES_COLL_NAME   = "uploaded-files"
	USER_DETAILS_COLL_NAME     = "user-details"
)

// Storage capacity constants
const (
	MONTHLY_STORAGE_ALLOCATION_SIZE = 50 // in gigabytes

	MONTHLY_STORAGE_SIZE        = 500  // in gigabytes
	FIXED_AMOUNT_1_STORAGE_SIZE = 1000 // in gigabytes
	FIXED_AMOUNT_2_STORAGE_SIZE = 2000 // in gigabytes
)

// User account types
const (
	MONTHLY_SUB    = "monthly"
	FIXED_AMOUNT_1 = "fa1"
	FIXED_AMOUNT_2 = "fa2"
)
