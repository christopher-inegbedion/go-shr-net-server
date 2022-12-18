package main

type UploadedFile struct {
	FileName        string     `json:"file_name"`
	FileSize        float64      `json:"file_size"`   // in gigabytes
	UploadDate      int        `json:"upload_date"` // in unix time
	InStoragePool   bool       `json:"in_storage_pool"`
	Hosts           [][]string `json:"hosts"`
	Shards          int        `json:"shards"`
	UploaderAddress string     `json:"uploader_address"`
	BackupShards    int        `json:"backup_shards"`
	IsMonthlySub    bool       `json:"is_monthly_sub"`
	Timezone        string     `json:"timezone"`
}
