package util

import (
	"github.com/eclipse-iofog/iofog-go-sdk/v3/pkg/client"
)

const LastSystemCatalogItemID int = 3

func IsSystemMsvc(msvc *client.MicroserviceInfo) bool {
	// 3 is hard coded. TODO: Find a way to maintain this ID from Controller.
	// Catalog item 1, 2, 3 are SYSTEM microservices, and are not inspectable by the user
	return msvc.CatalogItemID != 0 && msvc.CatalogItemID <= LastSystemCatalogItemID
}
