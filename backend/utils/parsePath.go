package utils

import (
	"strings"
	"path/filepath"
	"github.com/DuckOfTheBooBoo/web-gallery-app/backend/models"
)

func GenerateParentChildDir(userID uint, path string) []models.FolderChild {
	if len(path) == 0 {
		return []models.FolderChild{}
	}
	
	pathSlice := strings.Split(path, "/")
	var parentChildSlice []models.FolderChild
	base := ""

	// for i in range(len(dirs)):
    // parent = dirs[i]
    // if parent == "":
    //     parent = "/"
    // base = join(base, parent)
    // if base:
    //     parent = base
    // if i + 1 != len(dirs):
    //     child = dirs[i+1]
    //     print(f"{parent} -> {child}")

	for i := 0; i < len(pathSlice); i++ {
		parent := pathSlice[i]
		if parent == "" {
			parent = "/"
		}
		base = filepath.Join(base, parent)
		if base != "" {
			parent = base
		}

		if i + 1 != len(pathSlice) {
			child := pathSlice[i+1]
			

			parentChild := models.FolderChild{
				UserID: userID,
				Parent: parent,
				Child: child,
			}
			parentChildSlice = append(parentChildSlice, parentChild)
		}
	}

	return parentChildSlice
}