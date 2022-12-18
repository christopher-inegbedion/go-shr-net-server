package main

type NetworkStorageState struct {
	Name                    string `bson:"name"`
	NumberOfMonthlySubs     int64  `bson:"number_of_monthly_subs"`
	NumberOfFixedAmount1Subs int64  `bson:"number_of_fixed_amount1_subs"`
	NumberOfFixedAmount2Subs int64  `bson:"number_of_fixed_amount2_subs"`
	TotalAwsStorageSize     float64  `bson:"total_aws_storage_size"`  // in gigabytes
	TotalAwsStorageUsed     float64  `bson:"total_aws_storage_used"`  // in gigabytes
	TotalStoragePoolSize    float64  `bson:"total_storage_pool_size"` // in gigabytes
	TotalStoragePoolUsed    float64  `bson:"total_storage_pool_used"` // in gigabytes
}
