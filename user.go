package main

type User struct {
	Address           string `bson:"address"`
	RelayAddress      string `bson:"relay_address"`
	UserName          string `bson:"user_name"`
	Timezone          string `bson:"timezone"`
	AccountType       string `bson:"account_type"`
	SpoolCapacityUsed float64  `bson:"spool_capacity_used"` // in gigabytes
	AwsCapacityUsed   float64  `bson:"aws_capacity_used"`   // in gigabytes
	NumFilesUploaded  int    `bson:"number_of_files"`
}
