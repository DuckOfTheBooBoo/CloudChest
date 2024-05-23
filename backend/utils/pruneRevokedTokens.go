package utils

import (
	"fmt"
	"log"
	"gorm.io/gorm"
)

func PruneRevokedTokens(db *gorm.DB) {
	log.Println("Start pruning revoked tokens");
	result := db.Exec("DELETE FROM tokens WHERE expiration_date < NOW();")

	if result.Error != nil {
		log.Println(result.Error.Error())
		return
	}

	logMsg := fmt.Sprintf("Pruned %d revoked tokens", result.RowsAffected)
	log.Println(logMsg)
}
