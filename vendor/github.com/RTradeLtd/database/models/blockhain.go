package models

// Blockchain lists our available to different blockchains
// We do not store administration, or keys in the database.
type Blockchain struct {
	Name    string `gorm:"type:varchar(255);not null"`
	IPCPath string `gorm:"type:varchar(255);not null"`
	RPCPath string `gorm:"type:varchar(255);not null"`
}
